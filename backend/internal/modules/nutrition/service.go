package nutrition

import (
	"context"
	"errors"
	"time"

	"artGallery/internal/common"
	"artGallery/internal/modules/user"
	"artGallery/internal/modules/workout"
)

var (
	ErrNotConfigured = errors.New("ai not configured")
	ErrRateLimited   = errors.New("rate limited")
	ErrNoProfile     = errors.New("no profile")
)

const (
	maxPerHour   = 20 // messages per user per hour: protects the API quota
	historyTurns = 10 // how many past messages the model sees
)

const blockedReply = "I can't help with that one. I can help with nutrition questions like meals, protein, hydration or eating around your workouts. Could you rephrase?"

type ProfileFinder interface {
	FindByClerkID(ctx context.Context, clerkUserID string) (*user.UserProfile, error)
}

type Service struct {
	repo     Repository
	profiles ProfileFinder
	sessions workout.SessionRepository
	plans    workout.PlanRepository
	ai       AI // nil when GEMINI_API_KEY is not set
}

func NewService(repo Repository, profiles ProfileFinder, sessions workout.SessionRepository, plans workout.PlanRepository, ai AI) *Service {
	return &Service{repo: repo, profiles: profiles, sessions: sessions, plans: plans, ai: ai}
}

func (s *Service) Enabled() bool { return s.ai != nil }

func (s *Service) List(ctx context.Context, userID string) ([]Message, error) {
	return s.repo.Recent(ctx, userID, 50)
}

func (s *Service) Clear(ctx context.Context, userID string) error {
	return s.repo.Clear(ctx, userID)
}

// Send answers one user message. Flow:
//  1. rate limit  2. load profile + training data  3. build prompt (rules + context + recent chat)
//  4. call Gemini  5. save BOTH messages (so a failed call leaves no half-conversation).
func (s *Service) Send(ctx context.Context, userID string, req SendMessageRequest) (*Message, *Message, error) {
	if s.ai == nil {
		return nil, nil, ErrNotConfigured
	}
	started := time.Now().UTC()

	if n, err := s.repo.CountUserMessagesSince(ctx, userID, started.Add(-time.Hour)); err != nil {
		return nil, nil, err
	} else if n >= maxPerHour {
		return nil, nil, ErrRateLimited
	}

	profile, err := s.profiles.FindByClerkID(ctx, userID)
	if errors.Is(err, user.ErrNotFound) {
		return nil, nil, ErrNoProfile
	}
	if err != nil {
		return nil, nil, err
	}

	// Training context. If any of this fails the chat still works, just with less context.
	sessions, _ := s.sessions.AllCompleted(ctx, userID)
	var todayPlans []workout.WorkoutPlan
	if req.Day != "" {
		q := workout.PlanQuery{PaginationQuery: common.PaginationQuery{Page: 1, Limit: 3}, Day: workout.Weekday(req.Day)}
		todayPlans, _, _ = s.plans.List(ctx, userID, q)
	}

	history, err := s.repo.Recent(ctx, userID, historyTurns)
	if err != nil {
		return nil, nil, err
	}
	turns := make([]Turn, 0, len(history)+1)
	for _, m := range history {
		role := "user"
		if m.Role == RoleAssistant {
			role = "model"
		}
		if len(turns) == 0 && role == "model" {
			continue // a conversation must start with a user turn
		}
		turns = append(turns, Turn{Role: role, Text: m.Content})
	}
	turns = append(turns, Turn{Role: "user", Text: req.Content}) // must be the LAST turn

	system := systemPrompt + "\n\n" + buildContext(profile, sessions, todayPlans, started)

	reply, err := s.ai.Generate(ctx, system, turns)
	if errors.Is(err, ErrAIBlocked) {
		reply, err = blockedReply, nil
	}
	if err != nil {
		return nil, nil, err
	}

	userMsg := &Message{UserID: userID, Role: RoleUser, Content: req.Content, CreatedAt: started}
	botMsg := &Message{UserID: userID, Role: RoleAssistant, Content: reply, CreatedAt: time.Now().UTC()}
	if botMsg.CreatedAt.Before(userMsg.CreatedAt) {
		botMsg.CreatedAt = userMsg.CreatedAt
	}
	if err := s.repo.Save(ctx, userMsg, botMsg); err != nil {
		return nil, nil, err
	}
	return userMsg, botMsg, nil
}

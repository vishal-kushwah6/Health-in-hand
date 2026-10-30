package workout

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// SessionService records what the user ACTUALLY did. Rules it enforces:
//   - every exercise must be built-in or the user's own (archived ones are fine,
//     so a plan that contains an archived exercise can still be logged)
//   - the plan (if any) must belong to the user
//   - the date cannot be more than a day in the future
type SessionService struct {
	sessions  SessionRepository
	plans     PlanRepository
	exercises ExerciseRepository
}

func NewSessionService(sessions SessionRepository, plans PlanRepository, exercises ExerciseRepository) *SessionService {
	return &SessionService{sessions: sessions, plans: plans, exercises: exercises}
}

func (s *SessionService) Create(ctx context.Context, userID string, req CreateSessionRequest) (*WorkoutSession, error) {
	if req.Date.After(time.Now().Add(24 * time.Hour)) {
		return nil, ErrInvalidDate
	}

	// resolve exercise ids once; remember raw string -> id for the name snapshot
	rawToID := make(map[string]bson.ObjectID, len(req.Exercises))
	seen := make(map[bson.ObjectID]struct{}, len(req.Exercises))
	ids := make([]bson.ObjectID, 0, len(req.Exercises))
	for _, ex := range req.Exercises {
		id, err := bson.ObjectIDFromHex(ex.ExerciseID)
		if err != nil {
			return nil, ErrInvalidExercises
		}
		rawToID[ex.ExerciseID] = id
		if _, dup := seen[id]; !dup {
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
	}

	found, err := s.exercises.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	nameByID := make(map[bson.ObjectID]string, len(found))
	for _, e := range found {
		if e.OwnerID == "" || e.OwnerID == userID { // never accept someone else's custom exercise
			nameByID[e.ID] = e.Name
		}
	}
	if len(nameByID) != len(ids) {
		return nil, ErrInvalidExercises
	}
	names := make(map[string]string, len(rawToID))
	for raw, id := range rawToID {
		names[raw] = nameByID[id]
	}

	if req.WorkoutPlanID != "" {
		pid, err := bson.ObjectIDFromHex(req.WorkoutPlanID)
		if err != nil {
			return nil, ErrInvalidPlan
		}
		if _, err := s.plans.FindByID(ctx, userID, pid); err != nil {
			if errors.Is(err, ErrPlanNotFound) {
				return nil, ErrInvalidPlan
			}
			return nil, err
		}
	}

	sess, err := req.ToModel(userID, names)
	if err != nil {
		return nil, ErrInvalidExercises
	}
	if err := s.sessions.Create(ctx, &sess); err != nil {
		return nil, err
	}
	return &sess, nil
}

func (s *SessionService) Get(ctx context.Context, userID string, id bson.ObjectID) (*WorkoutSession, error) {
	return s.sessions.FindByID(ctx, userID, id)
}

func (s *SessionService) List(ctx context.Context, userID string, q SessionQuery) ([]WorkoutSession, int64, error) {
	return s.sessions.List(ctx, userID, q)
}

func (s *SessionService) Delete(ctx context.Context, userID string, id bson.ObjectID) error {
	return s.sessions.Delete(ctx, userID, id)
}

func (s *SessionService) LastSets(ctx context.Context, userID string, ids []bson.ObjectID) (map[string][]SetLog, error) {
	return s.sessions.LastSets(ctx, userID, ids)
}

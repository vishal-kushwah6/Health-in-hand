package workout

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ProgressService struct {
	sessions SessionRepository
}

func NewProgressService(sessions SessionRepository) *ProgressService {
	return &ProgressService{sessions: sessions}
}

func (s *ProgressService) Summary(ctx context.Context, userID string, tzMinutes int) (ProgressSummaryResponse, error) {
	sessions, err := s.sessions.AllCompleted(ctx, userID)
	if err != nil {
		return ProgressSummaryResponse{}, err
	}
	loc := time.FixedZone("user", tzMinutes*60)
	return computeSummary(sessions, loc, time.Now()), nil
}

func (s *ProgressService) Exercise(ctx context.Context, userID string, exerciseID bson.ObjectID, rng string) (ExerciseProgressResponse, error) {
	sessions, err := s.sessions.AllCompleted(ctx, userID)
	if err != nil {
		return ExerciseProgressResponse{}, err
	}
	var since time.Time
	now := time.Now()
	switch rng {
	case "week":
		since = now.AddDate(0, 0, -7)
	case "month":
		since = now.AddDate(0, 0, -30)
	case "three_months":
		since = now.AddDate(0, 0, -90)
	case "year":
		since = now.AddDate(0, 0, -365)
	}
	return computeExerciseProgress(sessions, exerciseID, since), nil
}

func (s *ProgressService) Records(ctx context.Context, userID string) ([]PersonalRecordResponse, error) {
	sessions, err := s.sessions.AllCompleted(ctx, userID)
	if err != nil {
		return nil, err
	}
	return computeRecords(sessions), nil
}

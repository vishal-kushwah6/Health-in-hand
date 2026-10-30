package workout

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// ExerciseService: the library is read-only; users can add, edit and archive
// their OWN custom exercises.
type ExerciseService struct {
	repo ExerciseRepository
}

func NewExerciseService(repo ExerciseRepository) *ExerciseService {
	return &ExerciseService{repo: repo}
}

func (s *ExerciseService) List(ctx context.Context, userID string, q ExerciseQuery) ([]Exercise, int64, error) {
	return s.repo.List(ctx, userID, q)
}

func (s *ExerciseService) Get(ctx context.Context, userID string, id bson.ObjectID) (*Exercise, error) {
	return s.repo.FindVisible(ctx, userID, id)
}

func (s *ExerciseService) Create(ctx context.Context, userID string, req CreateExerciseRequest) (*Exercise, error) {
	e := req.ToModel(userID)
	if err := s.repo.Create(ctx, &e); err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *ExerciseService) Update(ctx context.Context, userID string, id bson.ObjectID, req UpdateExerciseRequest) (*Exercise, error) {
	e, err := s.repo.FindVisible(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if e.OwnerID != userID { // visible but built-in
		return nil, ErrReadOnlyExercise
	}
	req.Apply(e)
	if err := s.repo.Update(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

// Archive hides a custom exercise from pickers. Old plans and sessions that
// reference it keep working.
func (s *ExerciseService) Archive(ctx context.Context, userID string, id bson.ObjectID) error {
	e, err := s.repo.FindVisible(ctx, userID, id)
	if err != nil {
		return err
	}
	if e.OwnerID != userID {
		return ErrReadOnlyExercise
	}
	return s.repo.Archive(ctx, userID, id)
}

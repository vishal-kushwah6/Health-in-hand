package workout

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// PlanService holds the plan rules. Its main job beyond CRUD: a plan may only
// contain exercises the user is allowed to use (built-in or their own).
type PlanService struct {
	plans     PlanRepository
	exercises ExerciseRepository
}

func NewPlanService(plans PlanRepository, exercises ExerciseRepository) *PlanService {
	return &PlanService{plans: plans, exercises: exercises}
}

// PlanDetail is a plan plus exercise names (id hex -> name) for display.
type PlanDetail struct {
	Plan  *WorkoutPlan
	Names map[string]string
}

func (s *PlanService) checkExercises(ctx context.Context, userID string, items []PlanExercise) error {
	seen := make(map[bson.ObjectID]struct{}, len(items))
	ids := make([]bson.ObjectID, 0, len(items))
	for _, it := range items {
		if _, dup := seen[it.ExerciseID]; !dup {
			seen[it.ExerciseID] = struct{}{}
			ids = append(ids, it.ExerciseID)
		}
	}
	found, err := s.exercises.FindVisibleByIDs(ctx, userID, ids)
	if err != nil {
		return err
	}
	if len(found) != len(ids) {
		return ErrInvalidExercises
	}
	return nil
}

func (s *PlanService) detail(ctx context.Context, p *WorkoutPlan) (*PlanDetail, error) {
	ids := make([]bson.ObjectID, 0, len(p.Exercises))
	for _, e := range p.Exercises {
		ids = append(ids, e.ExerciseID)
	}
	found, err := s.exercises.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	names := make(map[string]string, len(found))
	for _, e := range found {
		names[e.ID.Hex()] = e.Name
	}
	return &PlanDetail{Plan: p, Names: names}, nil
}

func (s *PlanService) Create(ctx context.Context, userID string, req CreatePlanRequest) (*PlanDetail, error) {
	p, err := req.ToModel(userID)
	if err != nil {
		return nil, ErrInvalidExercises
	}
	if err := s.checkExercises(ctx, userID, p.Exercises); err != nil {
		return nil, err
	}
	if err := s.plans.Create(ctx, &p); err != nil {
		return nil, err
	}
	return s.detail(ctx, &p)
}

func (s *PlanService) Get(ctx context.Context, userID string, id bson.ObjectID) (*PlanDetail, error) {
	p, err := s.plans.FindByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return s.detail(ctx, p)
}

func (s *PlanService) List(ctx context.Context, userID string, q PlanQuery) ([]WorkoutPlan, int64, error) {
	return s.plans.List(ctx, userID, q)
}

func (s *PlanService) Update(ctx context.Context, userID string, id bson.ObjectID, req UpdatePlanRequest) (*PlanDetail, error) {
	p, err := s.plans.FindByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if err := req.Apply(p); err != nil {
		return nil, ErrInvalidExercises
	}
	// Only re-check exercises when the client actually replaced them, so renaming
	// a plan still works after one of its custom exercises was archived.
	if req.Exercises != nil {
		if err := s.checkExercises(ctx, userID, p.Exercises); err != nil {
			return nil, err
		}
	}
	if err := s.plans.Update(ctx, p); err != nil {
		return nil, err
	}
	return s.detail(ctx, p)
}

func (s *PlanService) Delete(ctx context.Context, userID string, id bson.ObjectID) error {
	return s.plans.Delete(ctx, userID, id)
}

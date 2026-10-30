package workout

import (
	"artGallery/internal/common"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// =====================================================================
// Exercise DTOs
// =====================================================================

// CreateExerciseRequest is for admin/seed use (the library is shared).
type CreateExerciseRequest struct {
	Name         string        `json:"name"         validate:"required,notblank,min=2,max=80"`
	Description  string        `json:"description"  validate:"omitempty,max=500"`
	MuscleGroups []MuscleGroup `json:"muscleGroups" validate:"required,min=1,max=5,unique,dive,oneof=chest back shoulders biceps triceps forearms core quads hamstrings glutes calves full_body cardio"`
	Equipment    Equipment     `json:"equipment"    validate:"required,oneof=none bodyweight barbell dumbbell kettlebell cable machine band"`
	Difficulty   Difficulty    `json:"difficulty"   validate:"required,oneof=beginner intermediate advanced"`
	Instructions []string      `json:"instructions" validate:"omitempty,max=15,dive,notblank,max=300"`
}

func (r CreateExerciseRequest) ToModel(ownerID string) Exercise {
	return Exercise{
		Name: strings.TrimSpace(r.Name), Description: strings.TrimSpace(r.Description),
		MuscleGroups: r.MuscleGroups, Equipment: r.Equipment, Difficulty: r.Difficulty,
		Instructions: r.Instructions, OwnerID: ownerID, Archived: false, CreatedAt: time.Now().UTC(),
	}
}

// ExerciseQuery: GET /api/exercises?muscleGroup=chest&search=press&page=1
type ExerciseQuery struct {
	common.PaginationQuery
	MuscleGroup MuscleGroup `json:"muscleGroup" form:"muscleGroup" validate:"omitempty,oneof=chest back shoulders biceps triceps forearms core quads hamstrings glutes calves full_body cardio"`
	Equipment   Equipment   `json:"equipment"   form:"equipment"   validate:"omitempty,oneof=none bodyweight barbell dumbbell kettlebell cable machine band"`
	Difficulty  Difficulty  `json:"difficulty"  form:"difficulty"  validate:"omitempty,oneof=beginner intermediate advanced"`
	Search      string      `json:"search"      form:"search"      validate:"omitempty,max=50"`
	// all = library + mine (default), mine = only my custom ones, library = only the built-in ones
	Scope string `json:"scope" form:"scope" validate:"omitempty,oneof=all mine library"`
}

type ExerciseResponse struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Description  string        `json:"description,omitempty"`
	MuscleGroups []MuscleGroup `json:"muscleGroups"`
	Equipment    Equipment     `json:"equipment"`
	Difficulty   Difficulty    `json:"difficulty"`
	Instructions []string      `json:"instructions,omitempty"`
	IsCustom     bool          `json:"isCustom"`
}

func NewExerciseResponse(e *Exercise) ExerciseResponse {
	return ExerciseResponse{
		ID: e.ID.Hex(), Name: e.Name, Description: e.Description, MuscleGroups: e.MuscleGroups,
		Equipment: e.Equipment, Difficulty: e.Difficulty, Instructions: e.Instructions,
		IsCustom: e.OwnerID != "",
	}
}

// =====================================================================
// WorkoutPlan DTOs
// =====================================================================

type PlanExerciseInput struct {
	ExerciseID  string `json:"exerciseId"  validate:"required,objectid"`
	Sets        int    `json:"sets"        validate:"required,gte=1,lte=20"`
	Reps        int    `json:"reps"        validate:"required,gte=1,lte=100"`
	RestSeconds int    `json:"restSeconds" validate:"omitempty,gte=0,lte=600"`
	Notes       string `json:"notes"       validate:"omitempty,max=200"`
}

// CreatePlanRequest: POST /api/workouts
type CreatePlanRequest struct {
	Name          string              `json:"name"          validate:"required,notblank,min=3,max=80"`
	Description   string              `json:"description"   validate:"omitempty,max=300"`
	Goal          common.FitnessGoal  `json:"goal"          validate:"omitempty,oneof=weight_loss muscle_gain general_fitness strength endurance maintain_fitness"`
	Exercises     []PlanExerciseInput `json:"exercises"     validate:"required,min=1,max=30,dive"`
	ScheduledDays []Weekday           `json:"scheduledDays" validate:"omitempty,max=7,unique,dive,oneof=monday tuesday wednesday thursday friday saturday sunday"`
}

func (r CreatePlanRequest) ToModel(userID string) (WorkoutPlan, error) {
	exs, err := toPlanExercises(r.Exercises)
	if err != nil {
		return WorkoutPlan{}, err
	}
	now := time.Now().UTC()
	return WorkoutPlan{
		UserID: userID, Name: strings.TrimSpace(r.Name), Description: strings.TrimSpace(r.Description),
		Goal: r.Goal, Exercises: exs, ScheduledDays: r.ScheduledDays,
		CreatedAt: now, UpdatedAt: now,
	}, nil
}

// UpdatePlanRequest: PUT /api/workouts/:id. nil = leave unchanged.
// If Exercises / ScheduledDays are sent, they REPLACE the old list.
type UpdatePlanRequest struct {
	Name          *string             `json:"name"          validate:"omitempty,notblank,min=3,max=80"`
	Description   *string             `json:"description"   validate:"omitempty,max=300"`
	Goal          *common.FitnessGoal `json:"goal"          validate:"omitempty,oneof=weight_loss muscle_gain general_fitness strength endurance maintain_fitness"`
	Exercises     []PlanExerciseInput `json:"exercises"     validate:"omitempty,min=1,max=30,dive"`
	ScheduledDays []Weekday           `json:"scheduledDays" validate:"omitempty,max=7,unique,dive,oneof=monday tuesday wednesday thursday friday saturday sunday"`
}

func (r UpdatePlanRequest) Apply(p *WorkoutPlan) error {
	if r.Name != nil {
		p.Name = strings.TrimSpace(*r.Name)
	}
	if r.Description != nil {
		p.Description = strings.TrimSpace(*r.Description)
	}
	if r.Goal != nil {
		p.Goal = *r.Goal
	}
	if r.Exercises != nil {
		exs, err := toPlanExercises(r.Exercises)
		if err != nil {
			return err
		}
		p.Exercises = exs
	}
	if r.ScheduledDays != nil {
		p.ScheduledDays = r.ScheduledDays
	}
	p.UpdatedAt = time.Now().UTC()
	return nil
}

func toPlanExercises(in []PlanExerciseInput) ([]PlanExercise, error) {
	out := make([]PlanExercise, 0, len(in))
	for i, e := range in {
		id, err := bson.ObjectIDFromHex(e.ExerciseID)
		if err != nil {
			return nil, fmt.Errorf("exercises[%d].exerciseId: invalid id", i)
		}
		out = append(out, PlanExercise{
			ExerciseID: id, Sets: e.Sets, Reps: e.Reps,
			RestSeconds: e.RestSeconds, Notes: strings.TrimSpace(e.Notes),
		})
	}
	return out, nil
}

type PlanExerciseResponse struct {
	ExerciseID   string `json:"exerciseId"`
	ExerciseName string `json:"exerciseName,omitempty"` // filled by the service from the exercises collection
	Sets         int    `json:"sets"`
	Reps         int    `json:"reps"`
	RestSeconds  int    `json:"restSeconds,omitempty"`
	Notes        string `json:"notes,omitempty"`
}

type PlanResponse struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description,omitempty"`
	Goal          common.FitnessGoal     `json:"goal,omitempty"`
	Exercises     []PlanExerciseResponse `json:"exercises"`
	ScheduledDays []Weekday              `json:"scheduledDays,omitempty"`
	CreatedAt     time.Time              `json:"createdAt"`
	UpdatedAt     time.Time              `json:"updatedAt"`
}

// NewPlanResponse leaves exerciseName empty; pass names (id hex -> name) to fill it.
func NewPlanResponse(p *WorkoutPlan, names map[string]string) PlanResponse {
	exs := make([]PlanExerciseResponse, 0, len(p.Exercises))
	for _, e := range p.Exercises {
		id := e.ExerciseID.Hex()
		exs = append(exs, PlanExerciseResponse{
			ExerciseID: id, ExerciseName: names[id], Sets: e.Sets, Reps: e.Reps,
			RestSeconds: e.RestSeconds, Notes: e.Notes,
		})
	}
	return PlanResponse{
		ID: p.ID.Hex(), Name: p.Name, Description: p.Description, Goal: p.Goal,
		Exercises: exs, ScheduledDays: p.ScheduledDays, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

// =====================================================================
// WorkoutSession DTOs
// =====================================================================

// SetLogInput: a set needs reps OR a duration (so planks/runs work).
type SetLogInput struct {
	Reps            int     `json:"reps"            validate:"required_without=DurationSeconds,gte=0,lte=200"`
	WeightKg        float64 `json:"weightKg"        validate:"gte=0,lte=1000"`
	DurationSeconds int     `json:"durationSeconds" validate:"required_without=Reps,gte=0,lte=36000"`
	Completed       bool    `json:"completed"`
}

type SessionExerciseInput struct {
	ExerciseID string        `json:"exerciseId" validate:"required,objectid"`
	Sets       []SetLogInput `json:"sets"       validate:"required,min=1,max=30,dive"`
}

// CreateSessionRequest: POST /api/workout-sessions
// Sent once when the user taps "Complete Workout".
type CreateSessionRequest struct {
	WorkoutPlanID   string                 `json:"workoutPlanId"   validate:"omitempty,objectid"`
	Name            string                 `json:"name"            validate:"required,notblank,max=80"`
	Date            time.Time              `json:"date"            validate:"required"`
	Exercises       []SessionExerciseInput `json:"exercises"       validate:"required,min=1,max=30,dive"`
	DurationMinutes int                    `json:"durationMinutes" validate:"omitempty,gte=0,lte=600"`
	Completed       bool                   `json:"completed"`
	Notes           string                 `json:"notes"           validate:"omitempty,max=500"`
}

// ToModel converts ids and sets set numbers. exerciseNames (id hex -> name)
// must be looked up by the service first; it is stored as a snapshot.
// The service should also reject dates more than ~1 day in the future.
func (r CreateSessionRequest) ToModel(userID string, exerciseNames map[string]string) (WorkoutSession, error) {
	s := WorkoutSession{
		UserID: userID, Name: strings.TrimSpace(r.Name), Date: r.Date.UTC(),
		DurationMinutes: r.DurationMinutes, Completed: r.Completed,
		Notes: strings.TrimSpace(r.Notes), CreatedAt: time.Now().UTC(),
	}
	if r.WorkoutPlanID != "" {
		pid, err := bson.ObjectIDFromHex(r.WorkoutPlanID)
		if err != nil {
			return WorkoutSession{}, fmt.Errorf("workoutPlanId: invalid id")
		}
		s.WorkoutPlanID = &pid
	}
	for i, ex := range r.Exercises {
		eid, err := bson.ObjectIDFromHex(ex.ExerciseID)
		if err != nil {
			return WorkoutSession{}, fmt.Errorf("exercises[%d].exerciseId: invalid id", i)
		}
		sets := make([]SetLog, 0, len(ex.Sets))
		for n, st := range ex.Sets {
			sets = append(sets, SetLog{
				SetNumber: n + 1, Reps: st.Reps, WeightKg: st.WeightKg,
				DurationSeconds: st.DurationSeconds, Completed: st.Completed,
			})
		}
		s.Exercises = append(s.Exercises, SessionExercise{
			ExerciseID: eid, ExerciseName: exerciseNames[ex.ExerciseID], Sets: sets,
		})
	}
	return s, nil
}

// SessionQuery: GET /api/workout-sessions?from=2026-10-01&to=2026-10-31
type SessionQuery struct {
	common.PaginationQuery
	From string `json:"from" form:"from" validate:"omitempty,datetime=2006-01-02"`
	To   string `json:"to"   form:"to"   validate:"omitempty,datetime=2006-01-02"`
}

type SetLogResponse struct {
	SetNumber       int     `json:"setNumber"`
	Reps            int     `json:"reps"`
	WeightKg        float64 `json:"weightKg"`
	DurationSeconds int     `json:"durationSeconds,omitempty"`
	Completed       bool    `json:"completed"`
}

type SessionExerciseResponse struct {
	ExerciseID   string           `json:"exerciseId"`
	ExerciseName string           `json:"exerciseName"`
	Sets         []SetLogResponse `json:"sets"`
}

type SessionResponse struct {
	ID              string                    `json:"id"`
	WorkoutPlanID   string                    `json:"workoutPlanId,omitempty"`
	Name            string                    `json:"name"`
	Date            time.Time                 `json:"date"`
	Exercises       []SessionExerciseResponse `json:"exercises"`
	DurationMinutes int                       `json:"durationMinutes,omitempty"`
	Completed       bool                      `json:"completed"`
	Notes           string                    `json:"notes,omitempty"`
}

func NewSessionResponse(s *WorkoutSession) SessionResponse {
	resp := SessionResponse{
		ID: s.ID.Hex(), Name: s.Name, Date: s.Date, DurationMinutes: s.DurationMinutes,
		Completed: s.Completed, Notes: s.Notes,
		Exercises: make([]SessionExerciseResponse, 0, len(s.Exercises)),
	}
	if s.WorkoutPlanID != nil {
		resp.WorkoutPlanID = s.WorkoutPlanID.Hex()
	}
	for _, ex := range s.Exercises {
		sets := make([]SetLogResponse, 0, len(ex.Sets))
		for _, st := range ex.Sets {
			sets = append(sets, SetLogResponse(st))
		}
		resp.Exercises = append(resp.Exercises, SessionExerciseResponse{
			ExerciseID: ex.ExerciseID.Hex(), ExerciseName: ex.ExerciseName, Sets: sets,
		})
	}
	return resp
}

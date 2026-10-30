package workout

import "strings"

// UpdateExerciseRequest: PUT /api/exercises/:id (custom exercises only).
// nil / omitted = leave unchanged.
type UpdateExerciseRequest struct {
	Name         *string       `json:"name"         validate:"omitempty,notblank,min=2,max=80"`
	Description  *string       `json:"description"  validate:"omitempty,max=500"`
	MuscleGroups []MuscleGroup `json:"muscleGroups" validate:"omitempty,min=1,max=5,unique,dive,oneof=chest back shoulders biceps triceps forearms core quads hamstrings glutes calves full_body cardio"`
	Equipment    *Equipment    `json:"equipment"    validate:"omitempty,oneof=none bodyweight barbell dumbbell kettlebell cable machine band"`
	Difficulty   *Difficulty   `json:"difficulty"   validate:"omitempty,oneof=beginner intermediate advanced"`
	Instructions []string      `json:"instructions" validate:"omitempty,max=15,dive,notblank,max=300"`
}

func (r UpdateExerciseRequest) Apply(e *Exercise) {
	if r.Name != nil {
		e.Name = strings.TrimSpace(*r.Name)
	}
	if r.Description != nil {
		e.Description = strings.TrimSpace(*r.Description)
	}
	if r.MuscleGroups != nil {
		e.MuscleGroups = r.MuscleGroups
	}
	if r.Equipment != nil {
		e.Equipment = *r.Equipment
	}
	if r.Difficulty != nil {
		e.Difficulty = *r.Difficulty
	}
	if r.Instructions != nil {
		e.Instructions = r.Instructions
	}
}

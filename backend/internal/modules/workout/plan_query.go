package workout

import (
	"artGallery/internal/common"
	"time"
)

// PlanQuery: GET /api/workouts?day=monday&page=1
type PlanQuery struct {
	common.PaginationQuery
	Day Weekday `json:"day" form:"day" validate:"omitempty,oneof=monday tuesday wednesday thursday friday saturday sunday"`
}

// PlanSummaryResponse is the lightweight item used in lists.
type PlanSummaryResponse struct {
	ID            string             `json:"id"`
	Name          string             `json:"name"`
	Description   string             `json:"description,omitempty"`
	Goal          common.FitnessGoal `json:"goal,omitempty"`
	ExerciseCount int                `json:"exerciseCount"`
	ScheduledDays []Weekday          `json:"scheduledDays,omitempty"`
	UpdatedAt     time.Time          `json:"updatedAt"`
}

func NewPlanSummaryResponse(p *WorkoutPlan) PlanSummaryResponse {
	return PlanSummaryResponse{
		ID: p.ID.Hex(), Name: p.Name, Description: p.Description, Goal: p.Goal,
		ExerciseCount: len(p.Exercises), ScheduledDays: p.ScheduledDays, UpdatedAt: p.UpdatedAt,
	}
}

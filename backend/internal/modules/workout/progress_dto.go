package workout

import "time"

// Progress is DERIVED from workout_sessions; nothing is stored separately.

// ProgressQuery: GET /api/progress?tz=330
// tz = the user's offset from UTC in minutes (India = 330). Weeks, months and
// streaks are counted in the user's own calendar days, not the server's.
type ProgressQuery struct {
	TZ int `json:"tz" form:"tz" validate:"gte=-840,lte=840"`
}

// ExerciseProgressQuery: GET /api/progress/exercise?exerciseId=...&range=three_months
type ExerciseProgressQuery struct {
	ExerciseID string `json:"exerciseId" form:"exerciseId" validate:"required,objectid"`
	Range      string `json:"range"      form:"range"      validate:"omitempty,oneof=week month three_months year all"`
}

// WeeklyPoint is one bar in the "workouts per week" chart (weeks start on Monday).
type WeeklyPoint struct {
	WeekStart string  `json:"weekStart"` // YYYY-MM-DD
	Workouts  int     `json:"workouts"`
	VolumeKg  float64 `json:"volumeKg"`
}

// ProgressSummaryResponse: GET /api/progress
type ProgressSummaryResponse struct {
	TotalWorkouts     int           `json:"totalWorkouts"`
	WorkoutsThisWeek  int           `json:"workoutsThisWeek"`
	WorkoutsThisMonth int           `json:"workoutsThisMonth"`
	CurrentStreakDays int           `json:"currentStreakDays"`
	LongestStreakDays int           `json:"longestStreakDays"`
	TotalVolumeKg     float64       `json:"totalVolumeKg"` // sum of reps x weight over completed sets
	Weekly            []WeeklyPoint `json:"weekly"`        // last 8 weeks, oldest first
}

// ExerciseProgressPoint is one session's best result for an exercise.
type ExerciseProgressPoint struct {
	Date         time.Time `json:"date"`
	BestWeightKg float64   `json:"bestWeightKg"`
	BestReps     int       `json:"bestReps"` // reps at the best weight
	VolumeKg     float64   `json:"volumeKg"`
	Estimated1RM float64   `json:"estimated1rm"` // Epley: weight x (1 + reps / 30)
}

type ExerciseProgressResponse struct {
	ExerciseID   string                  `json:"exerciseId"`
	ExerciseName string                  `json:"exerciseName"`
	Points       []ExerciseProgressPoint `json:"points"` // oldest first
}

type PersonalRecordResponse struct {
	ExerciseID   string    `json:"exerciseId"`
	ExerciseName string    `json:"exerciseName"`
	WeightKg     float64   `json:"weightKg"`
	Reps         int       `json:"reps"`
	AchievedOn   time.Time `json:"achievedOn"`
}

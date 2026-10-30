package workout

import "time"

// SessionSummaryResponse is the lightweight item used in the history list.
// Only COMPLETED sets count towards sets and volume.
type SessionSummaryResponse struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Date            time.Time `json:"date"`
	ExerciseCount   int       `json:"exerciseCount"`
	SetCount        int       `json:"setCount"`
	VolumeKg        float64   `json:"volumeKg"` // sum of reps x weight
	DurationMinutes int       `json:"durationMinutes,omitempty"`
	Completed       bool      `json:"completed"`
}

func NewSessionSummaryResponse(s *WorkoutSession) SessionSummaryResponse {
	sets := 0
	volume := 0.0
	for _, ex := range s.Exercises {
		for _, st := range ex.Sets {
			if st.Completed {
				sets++
				volume += float64(st.Reps) * st.WeightKg
			}
		}
	}
	return SessionSummaryResponse{
		ID: s.ID.Hex(), Name: s.Name, Date: s.Date, ExerciseCount: len(s.Exercises),
		SetCount: sets, VolumeKg: volume, DurationMinutes: s.DurationMinutes, Completed: s.Completed,
	}
}

// LastSetsResponse: for each exercise id, the sets from the most recent session
// that included it. The app uses it to pre-fill weights ("last time: 40 kg x 10").
type LastSetResponse struct {
	Reps     int     `json:"reps"`
	WeightKg float64 `json:"weightKg"`
}

type LastSetsResponse struct {
	Items map[string][]LastSetResponse `json:"items"`
}

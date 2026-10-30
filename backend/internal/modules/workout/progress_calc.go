package workout

import (
	"math"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Pure functions: sessions in, numbers out. No database, easy to reason about.
// All inputs are COMPLETED sessions sorted oldest first.

// dayNumber turns an instant into a whole "calendar day" number in the user's
// timezone, so consecutive days differ by exactly 1.
func dayNumber(t time.Time, loc *time.Location) int {
	l := t.In(loc)
	d := time.Date(l.Year(), l.Month(), l.Day(), 0, 0, 0, 0, time.UTC)
	return int(d.Unix() / 86400)
}

func dayString(day int) string {
	return time.Unix(int64(day)*86400, 0).UTC().Format("2006-01-02")
}

func round1(v float64) float64 { return math.Round(v*10) / 10 }

// completedVolume = sum of reps x weight over completed sets.
func completedVolume(s *WorkoutSession) float64 {
	v := 0.0
	for _, ex := range s.Exercises {
		for _, st := range ex.Sets {
			if st.Completed {
				v += float64(st.Reps) * st.WeightKg
			}
		}
	}
	return v
}

// estimate1RM uses the Epley formula.
func estimate1RM(weight float64, reps int) float64 {
	if weight <= 0 || reps <= 0 {
		return 0
	}
	if reps == 1 {
		return weight
	}
	return weight * (1 + float64(reps)/30)
}

// computeSummary: totals, this week / month, streaks and the last 8 weeks.
//
// Streak = consecutive calendar days with at least one completed workout.
// The current streak stays alive if you have not trained YET today
// (it counts back from yesterday), and breaks once a full day is missed.
func computeSummary(sessions []WorkoutSession, loc *time.Location, now time.Time) ProgressSummaryResponse {
	today := dayNumber(now, loc)
	nowLocal := now.In(loc)
	weekStart := today - (int(nowLocal.Weekday())+6)%7 // Monday of this week

	resp := ProgressSummaryResponse{Weekly: make([]WeeklyPoint, 8)}
	for i := range resp.Weekly {
		resp.Weekly[i].WeekStart = dayString(weekStart - 7*(7-i))
	}

	days := make(map[int]bool)
	for i := range sessions {
		s := &sessions[i]
		d := dayNumber(s.Date, loc)
		days[d] = true

		vol := completedVolume(s)
		resp.TotalWorkouts++
		resp.TotalVolumeKg += vol

		if d >= weekStart && d <= today {
			resp.WorkoutsThisWeek++
		}
		sl := s.Date.In(loc)
		if sl.Year() == nowLocal.Year() && sl.Month() == nowLocal.Month() {
			resp.WorkoutsThisMonth++
		}

		back := 0
		if d < weekStart {
			back = (weekStart - d + 6) / 7 // whole weeks back, rounded up
		}
		if back <= 7 {
			resp.Weekly[7-back].Workouts++
			resp.Weekly[7-back].VolumeKg += vol
		}
	}
	resp.TotalVolumeKg = round1(resp.TotalVolumeKg)
	for i := range resp.Weekly {
		resp.Weekly[i].VolumeKg = round1(resp.Weekly[i].VolumeKg)
	}

	// current streak
	d := today
	if !days[d] {
		d-- // not trained yet today: streak is still alive through yesterday
	}
	for days[d] {
		resp.CurrentStreakDays++
		d--
	}

	// longest streak
	keys := make([]int, 0, len(days))
	for k := range days {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	run := 0
	for i, k := range keys {
		if i > 0 && k == keys[i-1]+1 {
			run++
		} else {
			run = 1
		}
		if run > resp.LongestStreakDays {
			resp.LongestStreakDays = run
		}
	}
	return resp
}

// computeExerciseProgress: one point per session that included the exercise.
// since.IsZero() means "all time".
func computeExerciseProgress(sessions []WorkoutSession, id bson.ObjectID, since time.Time) ExerciseProgressResponse {
	resp := ExerciseProgressResponse{ExerciseID: id.Hex(), Points: []ExerciseProgressPoint{}}
	for i := range sessions {
		s := &sessions[i]
		if !since.IsZero() && s.Date.Before(since) {
			continue
		}
		for _, ex := range s.Exercises {
			if ex.ExerciseID != id {
				continue
			}
			resp.ExerciseName = ex.ExerciseName
			var p ExerciseProgressPoint
			has := false
			for _, st := range ex.Sets {
				if !st.Completed {
					continue
				}
				has = true
				p.VolumeKg += float64(st.Reps) * st.WeightKg
				if st.WeightKg > p.BestWeightKg || (st.WeightKg == p.BestWeightKg && st.Reps > p.BestReps) {
					p.BestWeightKg, p.BestReps = st.WeightKg, st.Reps
				}
				if e := estimate1RM(st.WeightKg, st.Reps); e > p.Estimated1RM {
					p.Estimated1RM = e
				}
			}
			if has {
				p.Date = s.Date
				p.VolumeKg = round1(p.VolumeKg)
				p.Estimated1RM = round1(p.Estimated1RM)
				resp.Points = append(resp.Points, p)
			}
		}
	}
	return resp
}

// computeRecords: the heaviest completed set per exercise (more reps wins a tie),
// most recently achieved first. Bodyweight exercises (0 kg) rank by reps.
func computeRecords(sessions []WorkoutSession) []PersonalRecordResponse {
	best := make(map[bson.ObjectID]*PersonalRecordResponse)
	for i := range sessions {
		s := &sessions[i]
		for _, ex := range s.Exercises {
			for _, st := range ex.Sets {
				if !st.Completed || st.Reps < 1 {
					continue
				}
				cur := best[ex.ExerciseID]
				if cur == nil {
					best[ex.ExerciseID] = &PersonalRecordResponse{
						ExerciseID: ex.ExerciseID.Hex(), ExerciseName: ex.ExerciseName,
						WeightKg: st.WeightKg, Reps: st.Reps, AchievedOn: s.Date,
					}
					continue
				}
				cur.ExerciseName = ex.ExerciseName
				if st.WeightKg > cur.WeightKg || (st.WeightKg == cur.WeightKg && st.Reps > cur.Reps) {
					cur.WeightKg, cur.Reps, cur.AchievedOn = st.WeightKg, st.Reps, s.Date
				}
			}
		}
	}
	out := make([]PersonalRecordResponse, 0, len(best))
	for _, v := range best {
		out = append(out, *v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].AchievedOn.After(out[j].AchievedOn) })
	if len(out) > 50 {
		out = out[:50]
	}
	return out
}

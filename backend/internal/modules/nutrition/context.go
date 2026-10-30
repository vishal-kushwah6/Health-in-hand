package nutrition

import (
	"fmt"
	"math"
	"strings"
	"time"

	"artGallery/internal/modules/user"
	"artGallery/internal/modules/workout"
)

// estimateMaintenance returns approximate daily maintenance calories using the
// Mifflin-St Jeor equation times an activity factor. It is an ESTIMATE.
func estimateMaintenance(p *user.UserProfile) int {
	bmr := 10*p.WeightKg + 6.25*p.HeightCm - 5*float64(p.Age)
	switch p.Gender {
	case user.GenderMale:
		bmr += 5
	case user.GenderFemale:
		bmr -= 161
	default:
		bmr -= 78 // midpoint when gender is not specified
	}
	factor := map[user.ActivityLevel]float64{
		user.ActivitySedentary: 1.2,
		user.ActivityLight:     1.375,
		user.ActivityModerate:  1.55,
		user.ActivityHigh:      1.725,
		user.ActivityVeryHigh:  1.9,
	}[p.ActivityLevel]
	if factor == 0 {
		factor = 1.4
	}
	return int(math.Round(bmr*factor/50) * 50)
}

// proteinRange returns a typical grams-per-kg range for the user's goal.
func proteinRange(goal string) (lo, hi float64) {
	switch goal {
	case "muscle_gain":
		return 1.6, 2.2
	case "strength", "weight_loss":
		return 1.6, 2.0
	case "endurance":
		return 1.2, 1.6
	default:
		return 1.2, 1.8
	}
}

func daysAgo(from, now time.Time) string {
	d := int(now.Sub(from).Hours() / 24)
	switch d {
	case 0:
		return "today"
	case 1:
		return "yesterday"
	default:
		return fmt.Sprintf("%d days ago", d)
	}
}

// buildContext turns the user's stored data into a short text block the model can use.
func buildContext(p *user.UserProfile, sessions []workout.WorkoutSession, todayPlans []workout.WorkoutPlan, now time.Time) string {
	var b strings.Builder
	b.WriteString("USER CONTEXT (estimates, may be imperfect):\n")
	fmt.Fprintf(&b, "- Goal: %s\n", strings.ReplaceAll(string(p.FitnessGoal), "_", " "))
	fmt.Fprintf(&b, "- Age %d, gender %s, height %.0f cm, weight %.1f kg, activity level %s\n",
		p.Age, strings.ReplaceAll(string(p.Gender), "_", " "), p.HeightCm, p.WeightKg, strings.ReplaceAll(string(p.ActivityLevel), "_", " "))
	if p.DietPreference != "" {
		fmt.Fprintf(&b, "- Diet preference: %s (respect this strictly)\n", strings.ReplaceAll(string(p.DietPreference), "_", " "))
	}
	if p.Age < 18 {
		b.WriteString("- The user is UNDER 18: no calorie or weight-loss targets.\n")
	} else {
		fmt.Fprintf(&b, "- Estimated maintenance energy: about %d kcal/day\n", estimateMaintenance(p))
		lo, hi := proteinRange(string(p.FitnessGoal))
		fmt.Fprintf(&b, "- Typical protein range for this goal: %.1f-%.1f g/kg, about %.0f-%.0f g/day\n", lo, hi, lo*p.WeightKg, hi*p.WeightKg)
	}

	last7, last28 := 0, 0
	var last *workout.WorkoutSession
	for i := range sessions {
		s := &sessions[i]
		age := now.Sub(s.Date)
		if age <= 7*24*time.Hour {
			last7++
		}
		if age <= 28*24*time.Hour {
			last28++
		}
		last = s // sessions are oldest first, so the final one is the latest
	}
	fmt.Fprintf(&b, "- Training: %d workouts in the last 7 days, %d in the last 28 days.", last7, last28)
	if last != nil {
		fmt.Fprintf(&b, " Last workout: %q %s.", last.Name, daysAgo(last.Date, now))
	}
	b.WriteString("\n")

	if len(todayPlans) > 0 {
		names := make([]string, 0, len(todayPlans))
		for _, pl := range todayPlans {
			names = append(names, fmt.Sprintf("%s (%d exercises)", pl.Name, len(pl.Exercises)))
		}
		fmt.Fprintf(&b, "- Planned workout today: %s\n", strings.Join(names, ", "))
	} else {
		b.WriteString("- No workout is planned for today.\n")
	}
	return b.String()
}

package community

import (
	"context"
	"time"
)

// SeedCommunities creates the starter communities on first run so the app
// has something to join. Does nothing once any community exists.
func SeedCommunities(ctx context.Context, repo CommunityRepository) error {
	n, err := repo.CountAll(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	now := time.Now().UTC()
	mk := func(name string, cat Category, desc string) Community {
		return Community{Name: name, Category: cat, Description: desc, CreatedBy: "system", CreatedAt: now, UpdatedAt: now}
	}
	return repo.InsertMany(ctx, []Community{
		mk("Running", CategoryRunning, "Training plans, race reports and tips for runners of every pace."),
		mk("Gym", CategoryGym, "Everything about lifting, routines and gym life."),
		mk("Weight Loss", CategoryWeightLoss, "Share progress, swap tips and keep each other motivated."),
		mk("Muscle Building", CategoryMuscleBuilding, "Hypertrophy training, programs and progress check-ins."),
		mk("Calisthenics", CategoryCalisthenics, "Bodyweight strength: pull-ups, dips, muscle-ups and beyond."),
		mk("Yoga", CategoryYoga, "Flexibility, mobility and mindful movement."),
		mk("Powerlifting", CategoryPowerlifting, "Squat, bench, deadlift. PRs, meets and technique."),
		mk("Nutrition", CategoryNutrition, "Meals, macros and healthy eating ideas."),
	})
}

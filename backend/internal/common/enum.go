package common

// Enums shared by several modules.
// NOTE: struct tags need literal values, so the `oneof=...` lists in DTOs
// must be kept in sync with these constants by hand.

type FitnessGoal string

const (
	GoalWeightLoss     FitnessGoal = "weight_loss"
	GoalMuscleGain     FitnessGoal = "muscle_gain"
	GoalGeneralFitness FitnessGoal = "general_fitness"
	GoalStrength       FitnessGoal = "strength"
	GoalEndurance      FitnessGoal = "endurance"
	GoalMaintain       FitnessGoal = "maintain_fitness"
)

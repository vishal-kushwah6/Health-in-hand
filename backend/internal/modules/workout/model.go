package workout

import (
	"artGallery/internal/common"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Difficulty string

const (
	DifficultyBeginner     Difficulty = "beginner"
	DifficultyIntermediate Difficulty = "intermediate"
	DifficultyAdvanced     Difficulty = "advanced"
)

type MuscleGroup string

const (
	MuscleChest     MuscleGroup = "chest"
	MuscleBack      MuscleGroup = "back"
	MuscleShoulders MuscleGroup = "shoulders"
	MuscleBiceps    MuscleGroup = "biceps"
	MuscleTriceps   MuscleGroup = "triceps"
	MuscleForearms  MuscleGroup = "forearms"
	MuscleCore      MuscleGroup = "core"
	MuscleQuads     MuscleGroup = "quads"
	MuscleHamstring MuscleGroup = "hamstrings"
	MuscleGlutes    MuscleGroup = "glutes"
	MuscleCalves    MuscleGroup = "calves"
	MuscleFullBody  MuscleGroup = "full_body"
	MuscleCardio    MuscleGroup = "cardio"
)

type Equipment string

const (
	EquipNone       Equipment = "none"
	EquipBodyweight Equipment = "bodyweight"
	EquipBarbell    Equipment = "barbell"
	EquipDumbbell   Equipment = "dumbbell"
	EquipKettlebell Equipment = "kettlebell"
	EquipCable      Equipment = "cable"
	EquipMachine    Equipment = "machine"
	EquipBand       Equipment = "band"
)

type Weekday string

const (
	Monday    Weekday = "monday"
	Tuesday   Weekday = "tuesday"
	Wednesday Weekday = "wednesday"
	Thursday  Weekday = "thursday"
	Friday    Weekday = "friday"
	Saturday  Weekday = "saturday"
	Sunday    Weekday = "sunday"
)

// ---------------------------------------------------------------------
// Exercise: reusable library entry, shared by all users.
// Collection: "exercises".
// ---------------------------------------------------------------------
type Exercise struct {
	ID           bson.ObjectID `bson:"_id,omitempty"`
	Name         string        `bson:"name"`
	Description  string        `bson:"description,omitempty"`
	MuscleGroups []MuscleGroup `bson:"muscleGroups"`
	Equipment    Equipment     `bson:"equipment"`
	Difficulty   Difficulty    `bson:"difficulty"`
	Instructions []string      `bson:"instructions,omitempty"`
	OwnerID      string        `bson:"ownerId,omitempty"` // Clerk id; empty = global library
	Archived     bool          `bson:"archived"`          // "deleted" custom exercises are hidden, not removed
	CreatedAt    time.Time     `bson:"createdAt"`
}

// ---------------------------------------------------------------------
// WorkoutPlan: what the user INTENDS to do. Collection: "workout_plans".
// UserID is the Clerk user id.
// ---------------------------------------------------------------------
type PlanExercise struct {
	ExerciseID  bson.ObjectID `bson:"exerciseId"`
	Sets        int           `bson:"sets"`
	Reps        int           `bson:"reps"`
	RestSeconds int           `bson:"restSeconds,omitempty"`
	Notes       string        `bson:"notes,omitempty"`
}

type WorkoutPlan struct {
	ID            bson.ObjectID      `bson:"_id,omitempty"`
	UserID        string             `bson:"userId"`
	Name          string             `bson:"name"`
	Description   string             `bson:"description,omitempty"`
	Goal          common.FitnessGoal `bson:"goal,omitempty"`
	Exercises     []PlanExercise     `bson:"exercises"` // array order = display order
	ScheduledDays []Weekday          `bson:"scheduledDays,omitempty"`
	CreatedAt     time.Time          `bson:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt"`
}

// ---------------------------------------------------------------------
// WorkoutSession: what the user ACTUALLY did. Collection: "workout_sessions".
// Each set is logged separately ("40kg x 10", "40kg x 9", ...).
// ---------------------------------------------------------------------
type SetLog struct {
	SetNumber       int     `bson:"setNumber"`
	Reps            int     `bson:"reps"`
	WeightKg        float64 `bson:"weightKg"`
	DurationSeconds int     `bson:"durationSeconds,omitempty"` // for timed exercises (plank, run)
	Completed       bool    `bson:"completed"`
}

type SessionExercise struct {
	ExerciseID   bson.ObjectID `bson:"exerciseId"`
	ExerciseName string        `bson:"exerciseName"` // snapshot, so history survives renames
	Sets         []SetLog      `bson:"sets"`
}

type WorkoutSession struct {
	ID              bson.ObjectID     `bson:"_id,omitempty"`
	UserID          string            `bson:"userId"`
	WorkoutPlanID   *bson.ObjectID    `bson:"workoutPlanId,omitempty"` // nil = freestyle workout
	Name            string            `bson:"name"`
	Date            time.Time         `bson:"date"`
	Exercises       []SessionExercise `bson:"exercises"`
	DurationMinutes int               `bson:"durationMinutes,omitempty"`
	Completed       bool              `bson:"completed"`
	Notes           string            `bson:"notes,omitempty"`
	CreatedAt       time.Time         `bson:"createdAt"`
}

// ExerciseIndexes -> collection "exercises".
// Names are unique PER OWNER (case-insensitive) among non-archived exercises,
// so two users can both own "Hip Thrust Variant" and an archived name is reusable.
func ExerciseIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "ownerId", Value: 1}, {Key: "name", Value: 1}},
			Options: options.Index().
				SetName("owner_name_unique").
				SetUnique(true).
				SetCollation(&options.Collation{Locale: "en", Strength: 2}).
				SetPartialFilterExpression(bson.M{"archived": false}),
		},
		{Keys: bson.D{{Key: "muscleGroups", Value: 1}}},
	}
}

// PlanIndexes -> collection "workout_plans".
func PlanIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{Keys: bson.D{{Key: "userId", Value: 1}, {Key: "createdAt", Value: -1}}},
	}
}

// SessionIndexes -> collection "workout_sessions". The second index speeds
// up per-exercise progress queries.
func SessionIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{Keys: bson.D{{Key: "userId", Value: 1}, {Key: "date", Value: -1}}},
		{Keys: bson.D{{Key: "userId", Value: 1}, {Key: "exercises.exerciseId", Value: 1}, {Key: "date", Value: -1}}},
	}
}

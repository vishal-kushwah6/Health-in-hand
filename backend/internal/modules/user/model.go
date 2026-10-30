package user

import (
	"artGallery/internal/common"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Gender string

const (
	GenderMale           Gender = "male"
	GenderFemale         Gender = "female"
	GenderOther          Gender = "other"
	GenderPreferNotToSay Gender = "prefer_not_to_say"
)

type ActivityLevel string

const (
	ActivitySedentary ActivityLevel = "sedentary"
	ActivityLight     ActivityLevel = "light"
	ActivityModerate  ActivityLevel = "moderate"
	ActivityHigh      ActivityLevel = "high"
	ActivityVeryHigh  ActivityLevel = "very_high"
)

type DietPreference string

const (
	DietNonVegetarian DietPreference = "non_vegetarian"
	DietVegetarian    DietPreference = "vegetarian"
	DietEggetarian    DietPreference = "eggetarian"
	DietVegan         DietPreference = "vegan"
	DietPescatarian   DietPreference = "pescatarian"
)

// UserProfile is the application's own user record. Clerk handles login;
// ClerkUserID is the link between the two. Collection: "user_profiles".
type UserProfile struct {
	ID             bson.ObjectID      `bson:"_id,omitempty"`
	ClerkUserID    string             `bson:"clerkUserId"`
	Username       string             `bson:"username"`
	Name           string             `bson:"name"`
	ProfilePicture string             `bson:"profilePicture,omitempty"`
	Bio            string             `bson:"bio,omitempty"`
	Age            int                `bson:"age"`
	Gender         Gender             `bson:"gender"`
	HeightCm       float64            `bson:"heightCm"`
	WeightKg       float64            `bson:"weightKg"`
	FitnessGoal    common.FitnessGoal `bson:"fitnessGoal"`
	ActivityLevel  ActivityLevel      `bson:"activityLevel"`
	DietPreference DietPreference     `bson:"dietPreference,omitempty"`
	CreatedAt      time.Time          `bson:"createdAt"`
	UpdatedAt      time.Time          `bson:"updatedAt"`
}

// Indexes to create at startup. Both are unique: one profile per Clerk
// user, and no two users share a username.
func Indexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{Keys: bson.D{{Key: "clerkUserId", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "username", Value: 1}}, Options: options.Index().SetUnique(true)},
	}
}

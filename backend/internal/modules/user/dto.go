package user

import (
	"artGallery/internal/common"
	"strings"
	"time"
)

// CreateProfileRequest is the body of POST /api/profile (first-time setup).
type CreateProfileRequest struct {
	Username       string              `json:"username"       validate:"required,username"`
	Name           string              `json:"name"           validate:"required,notblank,min=2,max=60"`
	ProfilePicture string              `json:"profilePicture" validate:"omitempty,url,max=500"`
	Bio            string              `json:"bio"            validate:"omitempty,max=300"`
	Age            int                 `json:"age"            validate:"required,gte=13,lte=100"`
	Gender         Gender              `json:"gender"         validate:"required,oneof=male female other prefer_not_to_say"`
	HeightCm       float64             `json:"heightCm"       validate:"required,gte=100,lte=250"`
	WeightKg       float64             `json:"weightKg"       validate:"required,gte=30,lte=300"`
	FitnessGoal    *common.FitnessGoal `json:"fitnessGoal"    validate:"required,oneof=weight_loss muscle_gain general_fitness strength endurance maintain_fitness"`
	ActivityLevel  ActivityLevel       `json:"activityLevel"  validate:"required,oneof=sedentary light moderate high very_high"`
	DietPreference DietPreference      `json:"dietPreference" validate:"omitempty,oneof=non_vegetarian vegetarian eggetarian vegan pescatarian"`
}

// ToModel builds a new UserProfile. clerkUserID comes from the verified
// token, NEVER from the request body.
func (r CreateProfileRequest) ToModel(clerkUserID string) UserProfile {
	now := time.Now().UTC()
	return UserProfile{
		ClerkUserID:    clerkUserID,
		Username:       strings.ToLower(r.Username),
		Name:           strings.TrimSpace(r.Name),
		ProfilePicture: r.ProfilePicture,
		Bio:            strings.TrimSpace(r.Bio),
		Age:            r.Age,
		Gender:         r.Gender,
		HeightCm:       r.HeightCm,
		WeightKg:       r.WeightKg,
		FitnessGoal:    *r.FitnessGoal,
		ActivityLevel:  r.ActivityLevel,
		DietPreference: r.DietPreference,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// UpdateProfileRequest is the body of PUT /api/profile. Every field is a
// pointer so "not sent" (nil) is different from "sent as zero/empty".
type UpdateProfileRequest struct {
	Username       *string             `json:"username"       validate:"omitempty,username"`
	Name           *string             `json:"name"           validate:"omitempty,notblank,min=2,max=60"`
	ProfilePicture *string             `json:"profilePicture" validate:"omitempty,url,max=500"`
	Bio            *string             `json:"bio"            validate:"omitempty,max=300"`
	Age            *int                `json:"age"            validate:"omitempty,gte=13,lte=100"`
	Gender         *Gender             `json:"gender"         validate:"omitempty,oneof=male female other prefer_not_to_say"`
	HeightCm       *float64            `json:"heightCm"       validate:"omitempty,gte=100,lte=250"`
	WeightKg       *float64            `json:"weightKg"       validate:"omitempty,gte=30,lte=300"`
	FitnessGoal    *common.FitnessGoal `json:"fitnessGoal"    validate:"omitempty,oneof=weight_loss muscle_gain general_fitness strength endurance maintain_fitness"`
	ActivityLevel  *ActivityLevel      `json:"activityLevel"  validate:"omitempty,oneof=sedentary light moderate high very_high"`
	DietPreference *DietPreference     `json:"dietPreference" validate:"omitempty,oneof=non_vegetarian vegetarian eggetarian vegan pescatarian"`
}

// Apply copies only the fields that were sent onto an existing profile.
func (r UpdateProfileRequest) Apply(p *UserProfile) {
	if r.Username != nil {
		p.Username = strings.ToLower(*r.Username)
	}
	if r.Name != nil {
		p.Name = strings.TrimSpace(*r.Name)
	}
	if r.ProfilePicture != nil {
		p.ProfilePicture = *r.ProfilePicture
	}
	if r.Bio != nil {
		p.Bio = strings.TrimSpace(*r.Bio)
	}
	if r.Age != nil {
		p.Age = *r.Age
	}
	if r.Gender != nil {
		p.Gender = *r.Gender
	}
	if r.HeightCm != nil {
		p.HeightCm = *r.HeightCm
	}
	if r.WeightKg != nil {
		p.WeightKg = *r.WeightKg
	}
	if r.FitnessGoal != nil {
		p.FitnessGoal = *r.FitnessGoal
	}
	if r.ActivityLevel != nil {
		p.ActivityLevel = *r.ActivityLevel
	}
	if r.DietPreference != nil {
		p.DietPreference = *r.DietPreference
	}
	p.UpdatedAt = time.Now().UTC()
}

// ProfileResponse is what the API returns. Never return the model directly.
type ProfileResponse struct {
	ID             string             `json:"id"`
	ClerkUserID    string             `json:"clerkUserId"`
	Username       string             `json:"username"`
	Name           string             `json:"name"`
	ProfilePicture string             `json:"profilePicture,omitempty"`
	Bio            string             `json:"bio,omitempty"`
	Age            int                `json:"age"`
	Gender         Gender             `json:"gender"`
	HeightCm       float64            `json:"heightCm"`
	WeightKg       float64            `json:"weightKg"`
	FitnessGoal    common.FitnessGoal `json:"fitnessGoal"`
	ActivityLevel  ActivityLevel      `json:"activityLevel"`
	DietPreference DietPreference     `json:"dietPreference,omitempty"`
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
}

func NewProfileResponse(p *UserProfile) ProfileResponse {
	return ProfileResponse{
		ID: p.ID.Hex(), ClerkUserID: p.ClerkUserID, Username: p.Username, Name: p.Name,
		ProfilePicture: p.ProfilePicture, Bio: p.Bio, Age: p.Age, Gender: p.Gender,
		HeightCm: p.HeightCm, WeightKg: p.WeightKg, FitnessGoal: p.FitnessGoal,
		ActivityLevel: p.ActivityLevel, DietPreference: p.DietPreference,
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

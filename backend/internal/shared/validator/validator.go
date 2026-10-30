// Package validation wraps go-playground/validator so every module
// validates DTOs the same way and returns the same error shape.
//
//	go get github.com/go-playground/validator/v10
package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	validate      *validator.Validate
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_.]{3,30}$`)
)

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())

	// Report errors using the JSON field name ("heightCm"), not the Go name ("HeightCm").
	validate.RegisterTagNameFunc(func(f reflect.StructField) string {
		name := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return f.Name
		}
		return name
	})

	// objectid: string must be a valid 24-char hex MongoDB ObjectID.
	_ = validate.RegisterValidation("objectid", func(fl validator.FieldLevel) bool {
		_, err := bson.ObjectIDFromHex(fl.Field().String())
		return err == nil
	})

	// username: 3-30 chars, letters, digits, underscore, dot.
	_ = validate.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		return usernameRegex.MatchString(fl.Field().String())
	})

	// notblank: not empty after trimming whitespace ("   " is rejected).
	_ = validate.RegisterValidation("notblank", func(fl validator.FieldLevel) bool {
		return strings.TrimSpace(fl.Field().String()) != ""
	})
}

// FieldErrors maps a field path (e.g. "exercises[0].sets") to a message.
type FieldErrors map[string]string

func (e FieldErrors) Error() string {
	parts := make([]string, 0, len(e))
	for f, m := range e {
		parts = append(parts, f+": "+m)
	}
	return strings.Join(parts, "; ")
}

// Validate checks a struct. It returns nil when the struct is valid,
// otherwise a FieldErrors you can send back as a 400 response.
func Validate(s any) FieldErrors {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}
	verrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return FieldErrors{"_": err.Error()}
	}
	out := make(FieldErrors, len(verrs))
	for _, fe := range verrs {
		// Namespace looks like "CreatePlanRequest.exercises[0].sets"; drop the struct name.
		path := fe.Namespace()
		if i := strings.Index(path, "."); i >= 0 {
			path = path[i+1:]
		}
		out[path] = message(fe)
	}
	return out
}

func message(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required", "required_without", "required_with":
		return "is required"
	case "notblank":
		return "must not be blank"
	case "min":
		if fe.Kind() == reflect.String {
			return fmt.Sprintf("must be at least %s characters", fe.Param())
		}
		return fmt.Sprintf("must have at least %s items", fe.Param())
	case "max":
		if fe.Kind() == reflect.String {
			return fmt.Sprintf("must be at most %s characters", fe.Param())
		}
		return fmt.Sprintf("must have at most %s items", fe.Param())
	case "gte":
		return fmt.Sprintf("must be %s or more", fe.Param())
	case "lte":
		return fmt.Sprintf("must be %s or less", fe.Param())
	case "oneof":
		return "must be one of: " + strings.ReplaceAll(fe.Param(), " ", ", ")
	case "url":
		return "must be a valid URL"
	case "objectid":
		return "must be a valid id"
	case "username":
		return "must be 3-30 characters: letters, numbers, underscores, dots"
	case "unique":
		return "must not contain duplicates"
	default:
		return "is invalid (" + fe.Tag() + ")"
	}
}

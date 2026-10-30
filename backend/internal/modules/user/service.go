package user

import (
	"context"
	"errors"
)

// Service holds the business rules. It knows nothing about Gin or Mongo.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Get(ctx context.Context, clerkUserID string) (*UserProfile, error) {
	return s.repo.FindByClerkID(ctx, clerkUserID)
}

// Create makes the profile right after sign-up. One profile per Clerk user.
func (s *Service) Create(ctx context.Context, clerkUserID string, req CreateProfileRequest) (*UserProfile, error) {
	if _, err := s.repo.FindByClerkID(ctx, clerkUserID); err == nil {
		return nil, ErrAlreadyExists
	} else if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	p := req.ToModel(clerkUserID)
	if err := s.repo.Create(ctx, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// Update applies only the fields the client sent.
func (s *Service) Update(ctx context.Context, clerkUserID string, req UpdateProfileRequest) (*UserProfile, error) {
	p, err := s.repo.FindByClerkID(ctx, clerkUserID)
	if err != nil {
		return nil, err
	}
	req.Apply(p)
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

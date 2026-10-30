package community

import (
	"context"
	"errors"
	"time"

	"artGallery/internal/shared/database"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// CommunityService. Join, leave and create touch two collections (the member
// record and the memberCount counter), so they run inside a transaction:
// either both change or neither does.
type CommunityService struct {
	tx   database.TxManager
	repo CommunityRepository
}

func NewCommunityService(tx database.TxManager, repo CommunityRepository) *CommunityService {
	return &CommunityService{tx: tx, repo: repo}
}

// CommunityView is a community plus "am I a member?" for the caller.
type CommunityView struct {
	Community Community
	IsMember  bool
}

func (s *CommunityService) List(ctx context.Context, userID string, q CommunityQuery) ([]CommunityView, int64, error) {
	items, total, err := s.repo.List(ctx, userID, q)
	if err != nil {
		return nil, 0, err
	}
	ids := make([]bson.ObjectID, 0, len(items))
	for _, c := range items {
		ids = append(ids, c.ID)
	}
	member, err := s.repo.MemberOf(ctx, userID, ids)
	if err != nil {
		return nil, 0, err
	}
	out := make([]CommunityView, 0, len(items))
	for _, c := range items {
		out = append(out, CommunityView{Community: c, IsMember: member[c.ID]})
	}
	return out, total, nil
}

func (s *CommunityService) Get(ctx context.Context, userID string, id bson.ObjectID) (*CommunityView, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	isMember, err := s.repo.IsMember(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	return &CommunityView{Community: *c, IsMember: isMember}, nil
}

// Create makes the community and joins its creator as the first member.
func (s *CommunityService) Create(ctx context.Context, userID string, req CreateCommunityRequest) (*CommunityView, error) {
	c := req.ToModel(userID)
	c.MemberCount = 1
	err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.repo.Create(ctx, &c); err != nil {
			return err
		}
		return s.repo.AddMember(ctx, &CommunityMember{CommunityID: c.ID, UserID: userID, JoinedAt: time.Now().UTC()})
	})
	if err != nil {
		return nil, err
	}
	return &CommunityView{Community: c, IsMember: true}, nil
}

// Join is idempotent: joining twice does not change anything.
func (s *CommunityService) Join(ctx context.Context, userID string, id bson.ObjectID) (*CommunityView, error) {
	err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		if _, err := s.repo.FindByID(ctx, id); err != nil {
			return err
		}
		already, err := s.repo.IsMember(ctx, id, userID)
		if err != nil {
			return err
		}
		if already {
			return nil
		}
		if err := s.repo.AddMember(ctx, &CommunityMember{CommunityID: id, UserID: userID, JoinedAt: time.Now().UTC()}); err != nil {
			return err
		}
		return s.repo.IncMembers(ctx, id, 1)
	})
	if err != nil && !errors.Is(err, ErrAlreadyMember) { // lost a race with ourselves: already joined
		return nil, err
	}
	return s.Get(ctx, userID, id)
}

// Leave is idempotent too.
func (s *CommunityService) Leave(ctx context.Context, userID string, id bson.ObjectID) (*CommunityView, error) {
	err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		if _, err := s.repo.FindByID(ctx, id); err != nil {
			return err
		}
		removed, err := s.repo.RemoveMember(ctx, id, userID)
		if err != nil {
			return err
		}
		if removed {
			return s.repo.IncMembers(ctx, id, -1)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, userID, id)
}

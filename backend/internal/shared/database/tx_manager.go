package database

import "context"

// TxManager lets services run several repository calls atomically without
// knowing about MongoDB. *Mongo implements it.
//
// Example (join a community: insert member + increment counter):
//
//	err := s.tx.WithTx(ctx, func(ctx context.Context) error {
//	    if err := s.members.Insert(ctx, member); err != nil { return err }
//	    return s.communities.IncMemberCount(ctx, id, 1)
//	})
type TxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

var _ TxManager = (*Mongo)(nil)

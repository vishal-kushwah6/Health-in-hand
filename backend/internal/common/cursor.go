package common

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Cursor pagination for feeds (posts, comments). Unlike page/skip, it does
// not repeat or skip items when new posts arrive while the user scrolls.
//
// Usage in a repository (newest-first list):
//
//	filter := bson.M{"communityId": cid}
//	if q.Cursor != "" {
//	    c, err := DecodeCursor(q.Cursor)   // -> 400 on ErrInvalidCursor
//	    ...
//	    filter = bson.M{"$and": bson.A{filter, c.OlderThanFilter()}}
//	}
//	// sort {createdAt: -1, _id: -1}, fetch Limit+1 items.
//	// If you got Limit+1, there is a next page: drop the extra item and
//	// set NextCursor = EncodeCursor(Cursor{last.CreatedAt, last.ID.Hex()}).
var ErrInvalidCursor = errors.New("invalid cursor")

type Cursor struct {
	CreatedAt time.Time `json:"createdAt"`
	ID        string    `json:"id"` // ObjectID as hex
}

// EncodeCursor uses URL-safe base64 (no + / =), so it is safe in a query string.
func EncodeCursor(c Cursor) string {
	b, _ := json.Marshal(c)
	return base64.RawURLEncoding.EncodeToString(b)
}

func DecodeCursor(s string) (*Cursor, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, ErrInvalidCursor
	}
	var c Cursor
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, ErrInvalidCursor
	}
	if _, err := bson.ObjectIDFromHex(c.ID); err != nil {
		return nil, ErrInvalidCursor
	}
	return &c, nil
}

// OlderThanFilter matches documents that come AFTER this cursor in a
// newest-first list (createdAt desc, _id desc as tie-breaker).
func (c Cursor) OlderThanFilter() bson.M {
	oid, _ := bson.ObjectIDFromHex(c.ID)
	return bson.M{"$or": bson.A{
		bson.M{"createdAt": bson.M{"$lt": c.CreatedAt}},
		bson.M{"createdAt": c.CreatedAt, "_id": bson.M{"$lt": oid}},
	}}
}

// CursorQuery is embedded in query structs of cursor-paginated endpoints.
type CursorQuery struct {
	Cursor string `json:"cursor" form:"cursor" validate:"omitempty,max=300"`
	Limit  int    `json:"limit"  form:"limit"  validate:"omitempty,gte=1,lte=50"`
}

func (q *CursorQuery) Normalize() {
	if q.Limit < 1 {
		q.Limit = DefaultLimit
	}
}

// CursorResponse: NextCursor is empty when there are no more items.
type CursorResponse[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"nextCursor,omitempty"`
}

package nutrition

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message is one chat message between a user and the nutritionist.
// Collection: "nutrition_messages". UserID is the Clerk user id.
type Message struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	UserID    string        `bson:"userId"`
	Role      Role          `bson:"role"`
	Content   string        `bson:"content"`
	CreatedAt time.Time     `bson:"createdAt"`
}

func Indexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{Keys: bson.D{{Key: "userId", Value: 1}, {Key: "createdAt", Value: -1}}},
	}
}

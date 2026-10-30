package database

import (
	"context"
	"fmt"
	"log"
	"time"

	configs "artGallery/config"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// Mongo replaces the old Postgres struct. It holds the client (needed for
// sessions/transactions and for Disconnect) and the database handle that
// repositories use to get collections.
type Mongo struct {
	Client *mongo.Client
	DB     *mongo.Database
}

// ConnectMongo opens the connection pool and verifies it with a ping.
//
//	go get go.mongodb.org/mongo-driver/v2
func ConnectMongo(cfg *configs.Config) (*Mongo, error) {
	if cfg.Mongo.URI == "" {
		return nil, fmt.Errorf("MONGO_URI is not set")
	}

	// Pool tuning (same idea as SetMaxOpenConns / SetConnMaxLifetime before).
	opts := options.Client().
		ApplyURI(cfg.Mongo.URI).
		SetMaxPoolSize(25).
		SetMinPoolSize(5).
		SetMaxConnIdleTime(30 * time.Minute).
		SetServerSelectionTimeout(10 * time.Second)

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verify connection explicitly.
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}

	log.Println("✅ MongoDB connected successfully")
	return &Mongo{Client: client, DB: client.Database(cfg.Mongo.DBName)}, nil
}

// Collection is a small shortcut: m.Collection(database.ColUsers)
func (m *Mongo) Collection(name string) *mongo.Collection {
	return m.DB.Collection(name)
}

// Close should be deferred in main.go for a clean shutdown.
func (m *Mongo) Close(ctx context.Context) error {
	return m.Client.Disconnect(ctx)
}

// WithTx runs fn inside a MongoDB transaction (replaces the Postgres version).
//
// IMPORTANT: use the ctx that is passed INTO fn for every database call
// inside it. That ctx carries the transaction; if you use the outer ctx,
// the call runs outside the transaction.
//
// If fn returns an error the transaction is aborted, otherwise committed.
// Transactions need a replica set (MongoDB Atlas always is one; a plain
// local `mongod` is not).
func (m *Mongo) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	session, err := m.Client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(txCtx context.Context) (any, error) {
		return nil, fn(txCtx)
	})
	return err
}

// EnsureIndexes creates indexes for one collection. It is safe to call on
// every startup (existing identical indexes are left alone). Called from
// main.go with each module's index list, e.g.
//
//	database.EnsureIndexes(ctx, m.DB, database.ColUsers, user.Indexes())
//
// (Kept generic so this package does not import the modules, which would
// create an import cycle.)
func EnsureIndexes(ctx context.Context, db *mongo.Database, collection string, models []mongo.IndexModel) error {
	if len(models) == 0 {
		return nil
	}
	if _, err := db.Collection(collection).Indexes().CreateMany(ctx, models); err != nil {
		return fmt.Errorf("create indexes for %s: %w", collection, err)
	}
	return nil
}

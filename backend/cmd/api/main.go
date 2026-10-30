package main

import (
	configs "artGallery/config"
	"artGallery/internal/modules/community"
	"artGallery/internal/modules/nutrition"
	"artGallery/internal/modules/user"
	"artGallery/internal/modules/workout"
	"artGallery/internal/server"
	"artGallery/internal/shared/database"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func main() {
	// Load config
	cfg := configs.Load()

	// Connect DB
	m, err := database.ConnectMongo(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Create indexes (safe to run on every startup)
	if err := ensureIndexes(m); err != nil {
		log.Fatal(err)
	}

	// Fill the built-in exercise library on first run
	if err := workout.SeedExercises(context.Background(), workout.NewMongoExerciseRepository(m.DB)); err != nil {
		log.Fatal(err)
	}

	// Starter communities on first run
	if err := community.SeedCommunities(context.Background(), community.NewMongoCommunityRepository(m.DB)); err != nil {
		log.Fatal(err)
	}

	// Create server
	srv := server.New(cfg, m)

	// Run server in goroutine
	go func() {
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("server failed: ", err) // e.g. port already in use
		}
	}()

	// ---- Graceful Shutdown ----

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit // wait for signal

	log.Println("⚠️ Received shutdown signal")

	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Stop accepting requests first, then close the database
	if err := srv.Shutdown(ctx); err != nil {
		log.Println("Server forced to shutdown:", err)
	}
	if err := m.Close(ctx); err != nil {
		log.Println("MongoDB close error:", err)
	}

	log.Println("✅ Server exited gracefully")
}

// ensureIndexes registers every module's indexes. Add a line here whenever
// a new collection gets indexes.
func ensureIndexes(m *database.Mongo) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Day 1 change: exercise names are unique per owner now, not globally.
	// Drop the old global index if it exists (the error is ignored on purpose).
	_ = m.DB.RunCommand(ctx, bson.D{
		{Key: "dropIndexes", Value: database.ColExercises},
		{Key: "index", Value: "name_1"},
	}).Err()

	sets := []struct {
		collection string
		indexes    []mongo.IndexModel
	}{
		{database.ColUserProfiles, user.Indexes()},
		{database.ColExercises, workout.ExerciseIndexes()},
		{database.ColWorkoutPlans, workout.PlanIndexes()},
		{database.ColWorkoutSessions, workout.SessionIndexes()},
		{database.ColCommunities, community.CommunityIndexes()},
		{database.ColCommunityMembers, community.MemberIndexes()},
		{database.ColPosts, community.PostIndexes()},
		{database.ColPostLikes, community.LikeIndexes()},
		{database.ColComments, community.CommentIndexes()},
		{database.ColNutritionMessages, nutrition.Indexes()},
	}

	for _, s := range sets {
		if err := database.EnsureIndexes(ctx, m.DB, s.collection, s.indexes); err != nil {
			return err
		}
	}
	log.Println("✅ MongoDB indexes ready")
	return nil
}

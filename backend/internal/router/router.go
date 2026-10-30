package routes

import (
	configs "artGallery/config"
	"artGallery/internal/infrastructure/cloudinary"
	media "artGallery/internal/media/domain"
	"artGallery/internal/modules/community"
	"artGallery/internal/modules/nutrition"
	"artGallery/internal/modules/uploads"
	"artGallery/internal/modules/user"
	"artGallery/internal/modules/workout"
	"artGallery/internal/shared/database"
	"artGallery/internal/shared/middleware"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes is the single place where modules are wired together:
// repository -> service -> handler -> routes.
func RegisterRoutes(r *gin.Engine, cfg *configs.Config, db *database.Mongo) {
	// Public
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Everything under /api requires a valid Clerk session token
	api := r.Group("/api", middleware.ClerkAuth(cfg.Auth.Clerksecret))

	// --- User module ---
	userRepo := user.NewMongoRepository(db.DB)
	userSvc := user.NewService(userRepo)
	user.NewHandler(userSvc).RegisterRoutes(api)

	// --- Workout module: exercises ---
	exerciseRepo := workout.NewMongoExerciseRepository(db.DB)
	workout.NewExerciseHandler(workout.NewExerciseService(exerciseRepo)).RegisterRoutes(api)

	// --- Workout module: plans ---
	planRepo := workout.NewMongoPlanRepository(db.DB)
	workout.NewPlanHandler(workout.NewPlanService(planRepo, exerciseRepo)).RegisterRoutes(api)

	// --- Workout module: sessions (what the user actually did) ---
	sessionRepo := workout.NewMongoSessionRepository(db.DB)
	workout.NewSessionHandler(workout.NewSessionService(sessionRepo, planRepo, exerciseRepo)).RegisterRoutes(api)

	// --- Workout module: progress (derived from sessions) ---
	workout.NewProgressHandler(workout.NewProgressService(sessionRepo)).RegisterRoutes(api)

	// --- Community module: communities, posts, likes, comments ---
	communityRepo := community.NewMongoCommunityRepository(db.DB)
	postRepo := community.NewMongoPostRepository(db.DB)
	community.NewCommunityHandler(community.NewCommunityService(db, communityRepo)).RegisterRoutes(api)

	// --- Media (Cloudinary): optional, enabled only when CLOUDINARY_URL is set ---
	var store media.Storage
	if cfg.Media.URL == "" {
		log.Println("⚠️  image uploads DISABLED: CLOUDINARY_URL is empty")
	} else if cs, err := cloudinary.NewCloudinaryStorage(cfg.Media.URL); err != nil {
		log.Printf("⚠️  image uploads DISABLED: bad CLOUDINARY_URL: %v", err)
	} else {
		store = cs
		log.Println("✅ image uploads enabled (Cloudinary)")
	}
	uploads.NewHandler(store).RegisterRoutes(api)

	postSvc := community.NewPostService(db, postRepo, communityRepo, userRepo)
	if store != nil {
		postSvc.WithImages(store, community.CloudinaryPrefix(cfg.Media.URL))
	}
	community.NewPostHandler(postSvc).RegisterRoutes(api)

	// --- Nutrition: AI nutritionist (Gemini), optional until GEMINI_API_KEY is set ---
	var ai nutrition.AI
	if cfg.Gemini.APIKey == "" {
		log.Println("⚠️  AI nutritionist DISABLED: GEMINI_API_KEY is empty")
	} else {
		ai = nutrition.NewGeminiClient(cfg.Gemini.APIKey, cfg.Gemini.Model)
		log.Printf("✅ AI nutritionist enabled (model %s)", cfg.Gemini.Model)
	}
	nutrition.NewHandler(nutrition.NewService(nutrition.NewMongoRepository(db.DB), userRepo, sessionRepo, planRepo, ai)).RegisterRoutes(api)

	// --- Next modules (same 3 lines each) ---
	// workout plans / sessions / progress: same pattern
	// community: repo -> service -> handler -> RegisterRoutes(api)
}

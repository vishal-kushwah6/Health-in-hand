package configs

import (
	logger "artGallery/internal/shared/logging"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Mongo replaces the old Postgres "Database" block.
type Mongo struct {
	URI    string // full connection string, e.g. mongodb+srv://user:pass@cluster/
	DBName string
}

type Redis struct {
	HOST     string
	PORT     string
	PASSWORD string
	DB       int
}

// Gemini = the AI nutritionist. Optional: without GEMINI_API_KEY the chat answers "not configured".
type Gemini struct {
	APIKey string
	Model  string
}

type Media struct {
	URL string
}

type Config struct {
	Server Server
	Mongo  Mongo
	Auth   Auth
	Logger logger.Logger // Keep this if your logger is initialized elsewhere
	Redis  Redis
	Media  Media
	Gemini Gemini
}

type Auth struct {
	Clerksecret string
}

type Server struct {
	Port               string
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
	CORSAllowedOrigins []string
}

func Load() *Config {
	// This will try to load .env from the root of your project
	envPaths := []string{".env", "./.env", "../.env", "../../.env"}

	envLoaded := false
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			log.Printf("✅ Loaded environment variables from: %s", path)
			envLoaded = true
			break
		}
	}

	if !envLoaded {
		log.Println("⚠️ No .env file found in search paths. Using system environment variables.")
	}

	cfg := &Config{
		Server: Server{
			Port:               getEnv("PORT", "8080"),
			ReadTimeout:        mustParseDuration("READ_TIMEOUT", "5s"),
			WriteTimeout:       mustParseDuration("WRITE_TIMEOUT", "5s"),
			IdleTimeout:        mustParseDuration("IDLE_TIMEOUT", "60s"),
			CORSAllowedOrigins: getEnvSlice("CORS_ALLOWED_ORIGINS"),
		},
		Mongo: Mongo{
			URI:    requireEnv("MONGO_URI"),
			DBName: getEnv("MONGO_DB_NAME", "fitness_app"),
		},
		Auth: Auth{
			Clerksecret: requireEnv("CLERK_SECRET_KEY"),
		},
		// Redis and Cloudinary are OPTIONAL for now (getEnv instead of requireEnv)
		// so the app starts without them. Delete Redis entirely if you never use it.
		Redis: Redis{
			HOST:     getEnv("REDIS_HOST", ""),
			PORT:     getEnv("REDIS_PORT", ""),
			PASSWORD: "",
			DB:       0,
		},
		Media: Media{
			URL: getEnv("CLOUDINARY_URL", ""),
		},
		Gemini: Gemini{
			APIKey: getEnv("GEMINI_API_KEY", ""),
			Model:  getEnv("GEMINI_MODEL", "gemini-3.5-flash-lite"),
		},
	}
	fmt.Println(cfg.Server.CORSAllowedOrigins)

	return cfg
}

func requireEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("❌ Missing required environment variable: %s", key)
	}
	return val
}

func getEnv(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func mustParseDuration(key, fallback string) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		val = fallback
	}

	d, err := time.ParseDuration(val)
	if err != nil {
		log.Fatalf("❌ Invalid duration for %s: %v", key, err)
	}
	return d
}

func getEnvSlice(key string) []string {
	val := os.Getenv(key)
	if val == "" {
		return []string{}
	}

	// Split the string by commas
	items := strings.Split(val, ",")

	// Trim whitespace and quotes from each item to prevent CORS matching failures
	for i, item := range items {
		items[i] = strings.Trim(item, " \"'")
	}

	return items
}

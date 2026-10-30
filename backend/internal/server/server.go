package server

import (
	configs "artGallery/config"
	routes "artGallery/internal/router"
	"artGallery/internal/shared/database"
	"log"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	engine *gin.Engine
	srv    *http.Server
	cfg    *configs.Config
}

func New(cfg *configs.Config, db *database.Mongo) *Server {
	gin.ForceConsoleColor()
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// CORS is only needed for browser clients (Expo web). The native Android
	// app does not use it. Set CORS_ALLOWED_ORIGINS in .env to enable it,
	// e.g. CORS_ALLOWED_ORIGINS=http://localhost:8081
	if len(cfg.Server.CORSAllowedOrigins) > 0 {
		r.Use(cors.New(cors.Config{
			AllowOrigins:     cfg.Server.CORSAllowedOrigins,
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
		}))
	}

	// Clerk
	clerk.SetKey(cfg.Auth.Clerksecret)

	routes.RegisterRoutes(r, cfg, db)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return &Server{
		engine: r,
		srv:    srv,
		cfg:    cfg,
	}
}

func (s *Server) Start() error {
	log.Printf("🚀 Server started on %s\n", s.srv.Addr)
	return s.srv.ListenAndServe()
}

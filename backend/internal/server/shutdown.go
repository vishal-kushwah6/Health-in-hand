package server

import (
	"context"
	"log"
)

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("🛑 Shutting down server...")
	return s.srv.Shutdown(ctx)
}

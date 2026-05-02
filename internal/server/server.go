package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type UpvoteHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
}

type Server struct {
	router  *chi.Mux
	handler UpvoteHandler
}

func New(handler UpvoteHandler, addr string) *http.Server {
	s := &Server{
		router:  chi.NewRouter(),
		handler: handler,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return &http.Server{
		Addr:    addr,
		Handler: s.router,
	}
}

func (s *Server) setupMiddleware() {
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.RequestID)
}

func (s *Server) setupRoutes() {
	s.router.Post("/upvotes", s.handler.Create)
	s.router.Delete("/upvotes/{upvoteID}", s.handler.Delete)
	s.router.Get("/upvotes", s.handler.List)
	s.router.Get("/upvotes/{upvoteID}", s.handler.Get)
}

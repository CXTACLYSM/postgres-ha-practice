package internal

import (
	"github.com/CXTACLYSM/postgres-ha-practice/internal/di"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func InitRouter(handlers *di.Handlers, middlewares *di.Middlewares) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Group(func(r chi.Router) {
		r.Use(middlewares.Metrics.Handle)
		r.Get("/", handlers.Target.ServeHTTP)
	})
	r.Get("/metrics", handlers.Metrics.ServeHTTP)

	return r
}

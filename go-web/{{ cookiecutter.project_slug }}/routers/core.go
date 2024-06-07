package routers

import (
	"net/http"

	"{{ cookiecutter.project_name }}/routers/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func NewRouter(handlers []handlers.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Get("/health", Health)

	for _, handler := range handlers {
		handler.RegisterRoute(r)
	}

	// subroutne example
	// r.Route("/users", func(r chi.Router) {
	// r.Post("/", h)
	// })

	return r
}

package routers

import (
	"net/http"

	"{{ cookiecutter.project_name }}/routers/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/fx"
)

// AsRoute annotates the given function with the necessary group tag and types for
// registering it as a handler in the fx application.
//
// Parameters:
//   - f: the function to be annotated.
//
// Returns:
//   - any: the annotated function.
func AsRoute(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(handlers.Handler)),
		fx.ResultTags(`group:"handlers"`),
	)
}

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

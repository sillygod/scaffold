package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"{{ cookiecutter.project_name }}/db"
	"{{ cookiecutter.project_name }}/internal/app"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Refiner refines the input data from the request
// validate -> refine data -> database
type Refiner interface {
	refine(context.Context, *http.Request, *db.Queries) (interface{}, error)
}

type Composer interface {
	compose(context.Context, *db.Queries) (interface{}, error)
}

type Handler interface {
	handle() http.HandlerFunc
	RegisterRoute(*chi.Mux)
}

type RequestContext struct {
	q      *db.Queries
	logger *zap.SugaredLogger
}

// Flow define the basic process flow for an API endpoint
// it should be responsible for
// - validation
// - business logic (encapsulated in service)
func Flow(rctx RequestContext, refiner Refiner, f func(ctx context.Context, refinedData interface{}) (interface{}, error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var refinedData interface{}
		var err error

		ctx := r.Context()
		q := rctx.q

		if refiner != nil {
			refinedData, err = refiner.refine(ctx, r, q)
			if err != nil {
				app.RenderError(w, err)
				return
			}
		}

		// return the composer output
		data, err := f(ctx, refinedData)
		if err != nil {
			app.RenderError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(data); err != nil {
			app.RenderError(w, err)
		}
	}
}

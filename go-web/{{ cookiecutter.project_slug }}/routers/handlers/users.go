package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"{{ cookiecutter.project_name }}/db"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type UserCreationValidator struct {
	Model db.User `json:"-"`
	Name  string  `json:"name"`
}

func (u *UserCreationValidator) refine(ctx context.Context, r *http.Request, q *db.Queries) (interface{}, error) {
	// do validation
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		return nil, err
	}

	u.Model.Name = u.Name
	return q.CreateUser(ctx, u.Name)
}

type UserCreationComposer struct {
	Name string `json:"name"`
	Seed int    `json:"seed"`
}

func (u *UserCreationComposer) compose(ctx context.Context, q *db.Queries) (interface{}, error) {
	return u, nil
}

func NewUserHandler(conn *sql.DB, logger *zap.SugaredLogger) *UserHandler {

	return &UserHandler{
		q:        db.New(conn),
		logger:   logger,
		refiner:  &UserCreationValidator{},
		composer: &UserCreationComposer{},
	}
}

// UserHandler should implement the HandlerFunc interface
type UserHandler struct {
	q        *db.Queries
	logger   *zap.SugaredLogger
	refiner  *UserCreationValidator
	composer *UserCreationComposer
}

func (u *UserHandler) RegisterRoute(r *chi.Mux) {
	r.Post("/users", u.handle())
}

func (u *UserHandler) handle() http.HandlerFunc {
	rctx := RequestContext{u.q, u.logger}
	return Flow(rctx, u.refiner, func(ctx context.Context, refinedData interface{}) (interface{}, error) {
		user := refinedData.(db.User)

		// We can do anything between refinedData and composer
		u.composer.Name = user.Name
		u.composer.Seed = 1233

		return u.composer.compose(ctx, u.q)
	})
}

package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"{{ cookiecutter.project_name }}/db"
	"{{ cookiecutter.project_name }}/routers/schemas"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type UserCreationValidator struct {
	Model  db.User                   `json:"-"`
	Schema schemas.CreateUserRequest `json:"-"`
}

// refine validates and refines the user creation request.
//
// It takes the following parameters:
// - ctx: the context.Context object for the request.
// - r: the http.Request object containing the user creation request.
// - q: the db.Queries object for accessing the database.
//
// It returns an interface{} representing the created user and an error if any.
func (u *UserCreationValidator) refine(ctx context.Context, r *http.Request, q *db.Queries) (interface{}, error) {

	if err := json.NewDecoder(r.Body).Decode(&u.Schema); err != nil {
		return nil, err
	}

	validate := validator.New()
	if err := validate.Struct(u.Schema); err != nil {
		return nil, err
	}

	u.Model.Name = u.Schema.Name
	return q.CreateUser(ctx, u.Model.Name)
}

type UserCreationComposer struct {
	Name string `json:"name"`
	Seed int    `json:"seed"`
}

// compose composes the user creation response.
// Currently, it just returns the UserCreationComposer object.
// However, you can add more logic here if needed.
//
// Parameters:
// - ctx: the context.Context object for the request.
// - q: the db.Queries object for accessing the database.
//
// Returns:
// - interface{}: the composed UserCreationComposer object.
// - error: a nil error.
func (u *UserCreationComposer) compose(ctx context.Context, q *db.Queries) (interface{}, error) {
	return u, nil
}

func NewUserHandler(conn *pgx.Conn, logger *zap.SugaredLogger) *UserHandler {

	return &UserHandler{
		q:        db.New(conn),
		logger:   logger,
		refiner:  &UserCreationValidator{},
		composer: &UserCreationComposer{},
	}
}

// UserHandler is a struct that implements the HandlerFunc interface.
//
// It contains the necessary dependencies for handling user creation requests.
// The refiner is responsible for validating the user creation request,
// while the composer is responsible for composing the response.
//
// Fields:
// - q: the database queries interface
// - logger: the logger used for logging
// - refiner: the validator for user creation request
// - composer: the composer for user creation response
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

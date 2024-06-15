package tests

import (
    "{{ cookiecutter.project_name }}/config"
    "{{ cookiecutter.project_name }}/db"
    "{{ cookiecutter.project_name }}/internal/app"
    "{{ cookiecutter.project_name }}/routers"
    "{{ cookiecutter.project_name }}/routers/handlers"

	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
)

type UserHandlerTestSuite struct {
	suite.Suite
	r *chi.Mux
}

func (u *UserHandlerTestSuite) SetupTest() {
	fx.New(
		fx.Provide(config.NewViper),
		fx.Provide(app.NewLogger),
		fx.Provide(db.NewSqliteDB),
		fx.Provide(
			fx.Annotate(
				routers.NewRouter,
				fx.ParamTags(`group:"handlers"`),
			),
			routers.AsRoute(handlers.NewUserHandler)),
		fx.Invoke(func(r *chi.Mux, vp *viper.Viper) {
			u.r = r

			// TODO: think a way to handle this in a better way
			name := vp.GetString("sqlite.db")
			migrations, err := db.Migrate("sqlite://" + name)
			if err != nil {
				panic(err)
			}

			u.T().Logf("Applied migrations: %d", len(migrations.Applied))

		}),
	)

}

func (u *UserHandlerTestSuite) TestCreateUserWithValidBody() {
	reqBody := []byte(`{
	"name": "John Doe",
	"email": "song@test.com",
	"password": "!@SDGsjfe",
	"password_repeat": "!@SDGsjfe"
	}`)
	req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()
	u.r.ServeHTTP(w, req)

	// read the response
	resp := w.Result()
	defer resp.Body.Close()
	u.Equal(200, resp.StatusCode)

	content, _ := io.ReadAll(resp.Body)

	res := handlers.UserCreationComposer{}
	if err := json.Unmarshal(content, &res); err != nil {
		panic(err)
	}

	u.Equal("John Doe", res.Name)
}

func (u *UserHandlerTestSuite) TestCreateUserWithOutBody() {
	req := httptest.NewRequest("POST", "/users", nil)
	w := httptest.NewRecorder()
	u.r.ServeHTTP(w, req)

	// read the response
	resp := w.Result()
	defer resp.Body.Close()
	u.Equal(400, resp.StatusCode)

	content, _ := io.ReadAll(resp.Body)
	var errResp app.MyError
	json.Unmarshal(content, &errResp)
	u.Equal(app.ErrorCodeUnknown, errResp.Code)
	u.Equal("unknown error: EOF", errResp.Message)

}

func TestUserHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(UserHandlerTestSuite))
}

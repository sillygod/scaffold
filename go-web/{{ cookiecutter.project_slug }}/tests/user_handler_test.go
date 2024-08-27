package tests

import (
	"exampleproj/config"
	"exampleproj/db"
	"exampleproj/internal/app"
	"exampleproj/routers"
	"exampleproj/routers/handlers"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
)

type UserHandlerTestSuite struct {
	suite.Suite
	r     *chi.Mux
	fxApp *fx.App
}

func (u *UserHandlerTestSuite) NewTestDB(lc fx.Lifecycle, cfg *config.Config) *pgx.Conn {
	test_db := cfg.DB.NAME
	ctx := context.Background()
	cfg.DB.NAME = "postgres"

	masterConn, err := pgx.Connect(ctx, db.GetPostgresqlDSN(cfg))
	if err != nil {
		panic(err)
	}

	_, err = masterConn.Exec(ctx, "create database "+test_db)
	if err != nil {
		panic(err)
	}

	cfg.DB.NAME = test_db

	conn, err := pgx.Connect(ctx, db.GetPostgresqlDSN(cfg))
	if err != nil {
		panic(err)
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {

			if err := conn.Close(ctx); err != nil {
				return err
			}

			if _, err := masterConn.Exec(ctx, "drop database "+test_db); err != nil {
				return err
			}

			return masterConn.Close(ctx)
		},
	})

	return conn
}

func (u *UserHandlerTestSuite) SetupSuite() {
	// temporarily hardcode the test database name
	os.Setenv("DB_NAME", "test_db")

	u.fxApp = fx.New(
		fx.Provide(config.NewViper),
		fx.Provide(config.NewConfig),
		fx.Provide(app.NewLogger),
		fx.Provide(u.NewTestDB),
		fx.Provide(
			fx.Annotate(
				routers.NewRouter,
				fx.ParamTags(`group:"handlers"`),
			),
			routers.AsRoute(handlers.NewUserHandler)),
		fx.Invoke(func(r *chi.Mux, cfg *config.Config) {
			u.r = r

			engine := cfg.DB.ENGINE
			var dsn string

			if engine == config.SQLite {
				dsn = "sqlite://" + cfg.DB.NAME
			}

			if engine == config.Postgres {
				dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?search_path=public&sslmode=disable",
					cfg.DB.USER,
					cfg.DB.PASSWORD,
					cfg.DB.HOST,
					cfg.DB.PORT,
					cfg.DB.NAME,
				)
			}

			migrations, err := db.Migrate(dsn)
			if err != nil {
				panic(err)
			}

			u.T().Logf("Applied migrations: %d", len(migrations.Applied))

		}),
	)

	u.fxApp.Start(context.Background())

}

func (u *UserHandlerTestSuite) TearDownSuite() {
	u.fxApp.Stop(context.Background())
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

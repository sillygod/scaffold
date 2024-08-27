package main

import (
	"net/http"

	"exampleproj/config"
	"exampleproj/db"
	"exampleproj/internal/app"
	"exampleproj/routers"
	"exampleproj/routers/handlers"

	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Provide(
			app.NewHTTPServer,
			fx.Annotate(
				routers.NewRouter,
				fx.ParamTags(`group:"handlers"`),
			),

			// Register other routes here
			routers.AsRoute(handlers.NewUserHandler),
		),

		fx.Provide(config.NewConfig),
		fx.Provide(app.NewLogger),
		fx.Provide(db.NewPostgresqlDB),
		fx.Provide(config.NewViper),
		fx.Invoke(func(*http.Server) {}),
	).Run()
}

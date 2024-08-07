package main

import (
	"net/http"

	"{{ cookiecutter.project_name }}/config"
	"{{ cookiecutter.project_name }}/db"
	"{{ cookiecutter.project_name }}/internal/app"
	"{{ cookiecutter.project_name }}/routers"
	"{{ cookiecutter.project_name }}/routers/handlers"

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

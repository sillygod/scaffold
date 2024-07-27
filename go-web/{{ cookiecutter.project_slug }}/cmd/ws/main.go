package main

import (
	"{{ cookiecutter.project_name }}/cache"
	"{{ cookiecutter.project_name }}/config"
	"{{ cookiecutter.project_name }}/internal/app"
	"{{ cookiecutter.project_name }}/routers"
	"{{ cookiecutter.project_name }}/routers/handlers"
	"net/http"

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

			routers.AsRoute(handlers.NewWebsocketHandler),
		),

		fx.Provide(config.NewViper),
		fx.Provide(config.NewConfig),
		fx.Provide(cache.NewRedis),
		fx.Provide(app.NewLogger),
		fx.Invoke(func(*http.Server) {}),
	).Run()

}

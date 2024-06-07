package config

import (
	"strings"

	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
	"go.uber.org/fx"
)

func NewViper(lc fx.Lifecycle) *viper.Viper {
	vp := viper.New()

	vp.SetDefault("app.addr", "0.0.0.0")
	vp.SetDefault("app.port", "8080")
	vp.SetDefault("app.env", "local") // local, staging, prod
	vp.SetDefault("sqlite.db", ":memory:")

	replacer := strings.NewReplacer(".", "_")
	vp.SetEnvKeyReplacer(replacer)

	gotenv.Load()

	vp.AutomaticEnv()

	return vp
}

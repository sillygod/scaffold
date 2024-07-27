package config

import (
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type DBEngine string
type Env string

const (
	SQLite   DBEngine = "sqlite"
	MySQL    DBEngine = "mysql"
	Postgres DBEngine = "postgres"
)

const (
	Local   Env = "local"
	Staging Env = "staging"
	Prod    Env = "prod"
)

type Config struct {
	App struct {
		Addr string
		Port string
		Env  Env
	} `mapstructure:"app"`

	DB struct {
		ENGINE   DBEngine `mapstructure:"engine"`
		NAME     string   `mapstructure:"name"`
		HOST     string   `mapstructure:"host"`
		PORT     string   `mapstructure:"port"`
		USER     string   `mapstructure:"user"`
		PASSWORD string   `mapstructure:"password"`
	} `mapstructure:"db"`

	REDIS struct {
		ADDR string `mapstructure:"addr"`
		PORT string `mapstructure:"port"`
	} `mapstructure:"redis"`

	WEB3 struct {
		BLASTRPC_URL      string `mapstructure:"blastrpc_url"`
		BLASTSCAN_API_KEY string `mapstructure:"blastscan_api_key"`
		PYTH_API_HOST     string `mapstructure:"pyth_api_host"`
	}
}

func NewConfig(vp *viper.Viper) *Config {
	config := &Config{}
	if err := vp.Unmarshal(config); err != nil {
		panic(err)
	}
	return config
}

func NewViper(lc fx.Lifecycle) *viper.Viper {
	vp := viper.New()

	// NOTE: SetDefault or BindEnv to make viper take envs into account
	// when calling unmarshal
	vp.SetDefault("app.addr", "0.0.0.0")
	vp.SetDefault("app.port", "8080")
	vp.SetDefault("app.env", Local)
	vp.SetDefault("db.engine", SQLite)
	vp.SetDefault("db.host", "localhost")
	vp.SetDefault("db.port", "5432")
	vp.SetDefault("db.user", "postgres")
	vp.SetDefault("db.name", "song")
	vp.SetDefault("db.password", "postgres")
	vp.SetDefault("redis.addr", "redis")
	vp.SetDefault("redis.port", "6379")
	vp.SetDefault("web3.pyth_api_host", "https://hermes.pyth.network")

	replacer := strings.NewReplacer(".", "_")
	vp.SetEnvKeyReplacer(replacer)

	vp.AutomaticEnv()

	return vp
}

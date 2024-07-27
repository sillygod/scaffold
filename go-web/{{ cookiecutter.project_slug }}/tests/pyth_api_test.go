package tests

import (
	"{{ cookiecutter.project_name }}/config"
	"{{ cookiecutter.project_name }}/internal/app"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
)

type PythAPIClientTestSuite struct {
	suite.Suite
	fxApp  *fx.App
	client *app.PythAPIClient
}

func (p *PythAPIClientTestSuite) SetupSuite() {

	p.fxApp = fx.New(
		fx.Provide(config.NewViper),
		fx.Provide(config.NewConfig),
		fx.Provide(app.NewLogger),
		fx.Provide(),
		fx.Invoke(func(cfg *config.Config) {
			p.client = app.NewPythAPIClient(cfg.WEB3.PYTH_API_HOST)
		}))

	p.fxApp.Start(context.Background())
}

func (p *PythAPIClientTestSuite) TestGetPrice() {
	// btc e62df6c8b4a85fe1a67db44dc12de5db330f7ac66b72dc658afedf0f4a415b43
	// eth ff61491a931112ddf1bd8147cd1b641375f79f5825126d665480874634fd0ace
	res, err := p.client.GetLatestPrices([]string{"e62df6c8b4a85fe1a67db44dc12de5db330f7ac66b72dc658afedf0f4a415b43"})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Price: %+v\n", res.Parsed[0].Price)
}

func (p *PythAPIClientTestSuite) TearDownSuite() {
	p.fxApp.Stop(context.Background())
}

func TestPythAPIClientTestSuite(t *testing.T) {
	suite.Run(t, new(PythAPIClientTestSuite))
}

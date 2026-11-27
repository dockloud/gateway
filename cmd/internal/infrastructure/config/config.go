package config

import (
	"os"

	"github.com/dockloud/config/pkg/config"
	"github.com/dockloud/config/pkg/secret"
	model "github.com/dockloud/gateway/cmd/internal/domain/config"
)

type viperConfig struct {
	Service struct {
		Name         string   `mapstructure:"name"`
		ID           string   `mapstructure:"id"`
		Port         string   `mapstructure:"port"`
		Timeout      int      `mapstructure:"timeout"`
		PublicRoutes []string `mapstructure:"public_routes"`
	} `mapstructure:"service"`

	Dependencies map[string]struct {
		Name    string `mapstructure:"name"`
		Address string `mapstructure:"address"`
	} `mapstructure:"dependencies"`
}

type secrets struct {
	ServiceToken  string
	UserAuthToken string
}

func LoadMergedConfig() (*model.MergedConfig, error) {
	confHandler := config.NewViperHandler[viperConfig](
		os.Getenv("CONFIG_PATH"),
		os.Getenv("CONFIG_NAME"),
		"yaml",
	)

	conf, err := confHandler.GetConfig()
	if err != nil {
		return nil, err
	}
	secr, err := loadSecrets()
	if err != nil {
		return nil, err
	}

	return conf.ToMergedConfig(*secr), nil
}

func loadSecrets() (*secrets, error) {
	secHandler := secret.NewDockerHandler()

	servJWT, err := secHandler.ReadSecret("gateway_jwt_secret")
	if err != nil {
		return nil, err
	}
	authJWT, err := secHandler.ReadSecret("user_auth_jwt_secret")
	if err != nil {
		return nil, err
	}

	return &secrets{
		ServiceToken:  servJWT,
		UserAuthToken: authJWT,
	}, nil
}

func (c viperConfig) ToMergedConfig(secr secrets) *model.MergedConfig {
	deps := make(map[string]model.DependencyConfig, len(c.Dependencies))
	for _, d := range c.Dependencies {
		deps[d.Name] = model.DependencyConfig{
			Name:    d.Name,
			Address: d.Address,
		}
	}

	return &model.MergedConfig{
		Name:         c.Service.Name,
		ID:           c.Service.ID,
		Port:         c.Service.Port,
		Timeout:      c.Service.Timeout,
		PublicRoutes: c.Service.PublicRoutes,
		JWT: model.JWTConfig{
			ServiceSecret: []byte(secr.ServiceToken),
			UserAuthToken: secr.UserAuthToken,
		},
		Dependencies: deps,
	}
}

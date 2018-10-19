package service

import (
	"github.com/ViBiOh/auth/pkg/provider"
	"github.com/ViBiOh/httputils/pkg/logger"
)

type providerConfig struct {
	factory func(map[string]interface{}) (provider.Auth, error)
	config  map[string]interface{}
}

func initProvider(name string, factory func(map[string]interface{}) (provider.Auth, error), config map[string]interface{}) provider.Auth {
	auth, err := factory(config)
	if err != nil {
		logger.Error(`%+v`, err)
		return nil
	}

	return auth
}

func initProviders(providersConfig map[string]providerConfig) []provider.Auth {
	providers := make([]provider.Auth, 0, len(providersConfig))

	for name, conf := range providersConfig {
		if auth := initProvider(name, conf.factory, conf.config); auth != nil {
			logger.Info(`Provider for %s configured`, name)
			providers = append(providers, auth)
		}
	}

	return providers
}

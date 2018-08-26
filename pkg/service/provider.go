package service

import (
	"log"

	"github.com/ViBiOh/auth/pkg/provider"
)

type providerConfig struct {
	factory func(map[string]interface{}) (provider.Auth, error)
	config  map[string]interface{}
}

func initProvider(name string, factory func(map[string]interface{}) (provider.Auth, error), config map[string]interface{}) provider.Auth {
	auth, err := factory(config)
	if err != nil {
		log.Printf(`[auth] Error while initializing %s provider: %v`, name, err)
		return nil
	}

	return auth
}

func initProviders(providersConfig map[string]providerConfig) []provider.Auth {
	providers := make([]provider.Auth, 0, len(providersConfig))

	for name, conf := range providersConfig {
		if auth := initProvider(name, conf.factory, conf.config); auth != nil {
			log.Printf(`[auth] Provider for %s configured`, name)
			providers = append(providers, auth)
		}
	}

	return providers
}
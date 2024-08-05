package config

import "github.com/caarlos0/env/v11"

func LoadEnv() error {
	config := config{}

	if err := env.Parse(config.App); err != nil {
		return err
	}

	if err := env.Parse(config.GitHub); err != nil {
		return err
	}

	Config = config

	return nil
}

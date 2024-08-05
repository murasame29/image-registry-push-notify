package config

var Config config

type config struct {
	GitHub struct {
		ApplicationID  int64  `env:"GITHUB_APPLICATION_ID"`
		InstallationID int64  `env:"GITHUB_INSTALLATION_ID"`
		Username       string `env:"GITHUB_USERNAME"`
		CrtPath        string `env:"GITHUB_CRT_PATH"`
	}

	App struct {
		LogLevel   string `env:"LOG_LEVEL"`
		ConfigPath string `env:"CONFIG_PATH"`
	}
}

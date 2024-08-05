package updater

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/murasame29/image-registry-push-notify/sample-app/internal/model"
	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	// log-level
	// default debug
	// trace, debug, info, warn, error
	LogLevel string

	// required
	// github-token is personal access token
	GitHubApplicationID     int64
	GitHubAppInstallationID int64
	GitHubUsername          string
	GitHubAppCrtPath        string

	// required - input json only
	RegistryConfig []RegistryConfig
}

type RegistryConfig struct {
	// optional
	// e.g. regex: ^[0-9][a-z]$
	// e.g. latest
	AllowImageTag string `yaml:"allowImageTag"`
	DenyImageTag  string `yaml:"denyImageTag"`
	// e.g. 211125717884.dkr.ecr.ap-northeast-1.amazonaws.com/example/sample/sample-app/app
	// e.g. 123456789012.dkr.ecr.ap-northeast-1.amazonaws.com/*/$1/$2/$3
	RegitryURI string `yaml:"registryURI"`
	// e.g.github.com/murasame29/image-registry-push-notify/services/sample/sample-app/app/dev/overlays
	// e.g.github.com/murasame29/image-registry-push-notify/services/$1/$2/$3/$env/overlays
	GitHubRepository string `yaml:"gitHubRepository"`
	// e.g. 123456789012: dev
	// e.g. 234567890123: staging
	// e.g. 345678901234: prod
	Env map[string]string `yaml:"env"`
}

func (c *RegistryConfig) buildRepositoryName(event *model.ECRPushEvent) (string, error) {
	environment, ok := c.Env[event.Account]
	if !ok {
		return "", fmt.Errorf("environment not match")
	}

	// $1 ,$2 ..に対応対応するピースを探す

	repositoryName := c.GitHubRepository
	splitedRepository := strings.Split(c.GitHubRepository, "/")
	for i := range splitedRepository {
		if strings.Contains(splitedRepository[i], "$") {
			if splitedRepository[i] == "$env" {
				repositoryName = strings.ReplaceAll(repositoryName, "$env", environment)
				continue
			}

			// $1 ,$2 ..に対応対応するピースを埋める
		}
	}

	if strings.Contains(repositoryName, "$") {
		return "", fmt.Errorf("missing repository name. %s", repositoryName)
	}

	return repositoryName, nil
}

func (c *RegistryConfig) checkAllowTag(tag string) bool {
	tags := strings.Split(c.AllowImageTag, ":")
	if len(tags) == 2 {
		if !strings.Contains(tags[0], "regexp") {
			return false
		}
		reg, err := regexp.Compile(strings.TrimSpace(tags[1]))
		if err != nil {
			return false
		}
		return reg.Match([]byte(tag))
	} else if len(tags) == 1 {
		return strings.TrimSpace(tags[0]) == tag
	}

	return false
}

func (c *RegistryConfig) checkDenyTag(tag string) bool {
	tags := strings.Split(c.DenyImageTag, ":")
	if len(tags) == 2 {
		if !strings.Contains(tags[0], "regex") {
			return false
		}
		reg, err := regexp.Compile(strings.TrimSpace(tags[1]))
		if err != nil {
			return false
		}
		return reg.Match([]byte(tag))
	} else if len(tags) == 1 {
		return strings.TrimSpace(tags[0]) == tag
	}

	return false
}

func ParseConfig(path string) ([]RegistryConfig, error) {
	var config []RegistryConfig

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *AppConfig) parseConfig(event *model.ECRPushEvent) (*RegistryConfig, error) {
	registryURI := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", event.Account, event.Region)

	registryConfig, ok := filterRegistryConfigs(registryURI, event.Detail.RepositoryName, c.RegistryConfig)
	if !ok {
		return nil, fmt.Errorf("config dont match")
	}

	return registryConfig, nil
}

func filterRegistryConfigs(registryURI, repositoryName string, registryConfig []RegistryConfig) (*RegistryConfig, bool) {
	eventRepositoryPath := strings.Split(repositoryName, "/")
	for _, config := range registryConfig {
		if strings.Contains(config.RegitryURI, registryURI) {
			repositoryPath := strings.Split(strings.Trim(config.RegitryURI, registryURI), "/")
			if filterRegistryConfig(repositoryPath, eventRepositoryPath) {
				return &config, true
			}
		}
	}
	return nil, false
}

func filterRegistryConfig(repositoryPath, eventRepositoryPath []string) bool {
	for i := range len(repositoryPath) {
		if repositoryPath[i] == "*" || strings.Contains(repositoryPath[i], "$") {
			continue
		}

		if repositoryPath[i] != eventRepositoryPath[i] {
			return false
		}
	}

	return true
}

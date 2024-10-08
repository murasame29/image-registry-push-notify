package updater

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/murasame29/image-registry-push-notify/sample-app/internal/model"
	"gopkg.in/yaml.v2"
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
	GitHubRepository string `yaml:"githubRepository"`
	// e.g. ap-northeast-1
	Region string `yaml:"region"`
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

	variableMap := make(map[string]string)

	// $1 ,$2 ..に対応対応するピースを探す
	registryName := removeEmpty(strings.Split(c.RegitryURI, "/"))
	splitedRegistryName := removeEmpty(strings.Split(event.Detail.RepositoryName, "/"))
	for i := range registryName {
		if strings.Contains(registryName[i], "*") {
			continue
		}
		if strings.Contains(registryName[i], "$") {
			variableMap[registryName[i]] = splitedRegistryName[i]
		}
	}

	repositoryName := c.GitHubRepository
	splitedRepository := strings.Split(c.GitHubRepository, "/")
	for i := range splitedRepository {
		if strings.Contains(splitedRepository[i], "$") {
			if splitedRepository[i] == "$env" {
				repositoryName = strings.ReplaceAll(repositoryName, "$env", environment)
				continue
			}

			for variable, value := range variableMap {
				repositoryName = strings.ReplaceAll(repositoryName, variable, value)
			}
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
		if !strings.Contains(tags[0], "regexp") {
			return false
		}
		reg, err := regexp.Compile(strings.TrimSpace(tags[1]))
		if err != nil {
			return false
		}
		return !reg.Match([]byte(tag))
	} else if len(tags) == 1 {
		return strings.TrimSpace(tags[0]) != tag
	}

	return false
}

func NewConfigWithFile(path string) ([]RegistryConfig, error) {
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
	registryConfig, ok := filterRegistryConfigs(event, c.RegistryConfig)
	if !ok {
		return nil, fmt.Errorf("config dont match")
	}

	return registryConfig, nil
}

func filterRegistryConfigs(event *model.ECRPushEvent, registryConfig []RegistryConfig) (*RegistryConfig, bool) {
	eventRepositoryPath := strings.Split(event.Detail.RepositoryName, "/")

	for _, config := range registryConfig {
		if config.Region != event.Region {
			continue
		}
		if _, ok := config.Env[event.Account]; !ok {
			continue
		}
		repositoryPath := strings.Split(config.RegitryURI, "/")
		if filterRegistryConfig(repositoryPath, eventRepositoryPath) {
			return &config, true
		}
	}
	return nil, false
}

func filterRegistryConfig(repositoryPath, eventRepositoryPath []string) bool {
	repositoryPath = removeEmpty(repositoryPath)
	eventRepositoryPath = removeEmpty(eventRepositoryPath)
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

func removeEmpty(in []string) []string {
	var result []string
	for i := range in {
		if len(strings.TrimSpace(in[i])) == 0 {
			continue
		}
		result = append(result, in[i])
	}

	return result
}

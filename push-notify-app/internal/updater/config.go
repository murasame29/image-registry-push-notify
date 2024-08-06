package updater

import (
	"fmt"
	"log"
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
	registryName := trimSpace(strings.Split(c.RegitryURI, "/"))
	splitedRegistryName := trimSpace(strings.Split(event.Detail.RepositoryName, "/"))
	log.Println(registryName, splitedRegistryName, len(registryName), len(splitedRegistryName))
	for i := range registryName {
		if strings.Contains(registryName[i], "*") {
			continue
		}
		log.Println(registryName[i])
		if strings.Contains(registryName[i], "$") {
			log.Println(registryName[i], splitedRegistryName[i])
			variableMap[registryName[i]] = splitedRegistryName[i]
		}
	}

	log.Println(variableMap)

	repositoryName := c.GitHubRepository
	splitedRepository := strings.Split(c.GitHubRepository, "/")
	for i := range splitedRepository {
		if strings.Contains(splitedRepository[i], "$") {
			if splitedRepository[i] == "$env" {
				repositoryName = strings.ReplaceAll(repositoryName, "$env", environment)
				continue
			}
			value, ok := variableMap[splitedRepository[i]]
			if !ok {
				log.Println("variable not found", splitedRepository[i])
				continue
			}

			repositoryName = strings.ReplaceAll(repositoryName, splitedRepository[i], value)

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
		log.Println(config)
		if config.Region != event.Region {
			log.Println("region not match")
			continue
		}
		if _, ok := config.Env[event.Account]; !ok {
			log.Println("account not match")
			continue
		}
		repositoryPath := strings.Split(config.RegitryURI, "/")
		if filterRegistryConfig(repositoryPath, eventRepositoryPath) {
			return &config, true
		}
		log.Println("filter not match")
	}
	return nil, false
}

func filterRegistryConfig(repositoryPath, eventRepositoryPath []string) bool {
	repositoryPath = trimSpace(repositoryPath)
	eventRepositoryPath = trimSpace(eventRepositoryPath)

	for i := range len(repositoryPath) {
		if repositoryPath[i] == "*" || strings.Contains(repositoryPath[i], "$") {
			log.Println("wildcard", repositoryPath[i])
			continue
		}

		if repositoryPath[i] != eventRepositoryPath[i] {
			log.Println("not match", repositoryPath[i], eventRepositoryPath[i])
			return false
		}
	}

	return true
}

func trimSpace(in []string) []string {
	var r []string
	for i := range in {
		if len(strings.TrimSpace(in[i])) == 0 {
			continue
		}
		r = append(r, in[i])
	}
	return r
}

package updater

import (
	"fmt"

	"github.com/murasame29/image-registry-push-notify/sample-app/internal/git"
	"github.com/murasame29/image-registry-push-notify/sample-app/internal/model"
)

func Update(config *AppConfig, event *model.ECRPushEvent) error {
	_, err := git.NewGitHub(config.GitHubApplicationID, config.GitHubAppInstallationID, config.GitHubUsername, config.GitHubAppCrtPath)
	if err != nil {
		return fmt.Errorf("failed to new github. error: %v", err)
	}

	regitryConfig, err := config.parseConfig(event)
	if err != nil {
		return fmt.Errorf("failed to new github. error: %v", err)
	}

	if !regitryConfig.checkAllowTag(event.Detail.ImageTag) {
		return fmt.Errorf("tag not allowed. tag: %s", event.Detail.ImageTag)
	}

	if !regitryConfig.checkDenyTag(event.Detail.ImageTag) {
		return fmt.Errorf("tag is deny. tag: %s", event.Detail.ImageTag)
	}

	return nil
}

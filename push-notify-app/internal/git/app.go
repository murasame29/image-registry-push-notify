package git

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v63/github"
)

func NewGitHubApp(
	ctx context.Context,
	applicationID,
	installationID int64,
	crtPath string,
) (*github.Client, string, error) {
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, applicationID, installationID, crtPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to new key from file: %v", err)
	}

	token, err := itr.Token(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get token: %v", err)
	}

	return github.NewClient(&http.Client{Transport: itr, Timeout: 5 * time.Second}), token, nil
}

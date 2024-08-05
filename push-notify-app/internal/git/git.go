package git

import (
	"context"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v63/github"
	"github.com/murasame29/image-registry-push-notify/sample-app/internal/log"
)

type GitHub struct {
	// token is private access token
	token    string
	username string

	autherName  string
	autherEmail string

	clinet *github.Client
}

func NewGitHub(applicationID, installID int64, username, crtPath string) (*GitHub, error) {
	client, token, err := NewGitHubApp(context.Background(), applicationID, installID, crtPath)
	if err != nil {
		return nil, err
	}
	return &GitHub{
		token:    token,
		username: username,

		autherName: username,

		clinet: client,
	}, nil
}

func (g *GitHub) Clone(ctx context.Context, repository string) (*git.Repository, error) {
	repo, err := git.PlainClone("/tmp", false, &git.CloneOptions{
		URL: repository,
		Auth: &http.BasicAuth{
			Username: g.username,
			Password: g.token,
		},
	})
	if err != nil {
		log.Error(ctx, "failed to clone repository. repository: %s error: %v", repository, err)
		return nil, err
	}

	return repo, nil
}

func (g *GitHub) Commit(ctx context.Context, repo *git.Repository, path string, message string) (string, error) {
	workspace, err := repo.Worktree()
	if err != nil {
		log.Error(ctx, "failed to open worktree. error: %v", err)
		return "", err
	}

	log.Info(ctx, "trying git add. path: %s", path)
	hash, err := workspace.Add(path)
	if err != nil {
		log.Error(ctx, "failed to git add. path: %s error: %v", path, err)
		return "", err
	}

	log.Info(ctx, "git add successfuly path:%s hash: %s", path, hash.String())
	status, err := workspace.Status()
	if err != nil {
		log.Error(ctx, "failed to git status. error: %v", err)
		return "", err
	}

	log.Info(ctx, status.String())

	log.Info(ctx, "trying git commit -m %s. hash: %s", message, hash.String())
	commit, err := workspace.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  g.autherName,
			Email: g.autherEmail,
			When:  time.Now(),
		},
	})

	if err != nil {
		log.Error(ctx, "failed to git commit. error: %v", err)
	}

	log.Info(ctx, "git commit successfuly hash: %s", commit.String())
	return commit.String(), nil
}

func (g *GitHub) Push(ctx context.Context, repo *git.Repository) error {
	log.Info(ctx, "trying reposiotry push to origin")
	if err := repo.PushContext(ctx, &git.PushOptions{
		Auth: &http.BasicAuth{
			Username: g.username,
			Password: g.token,
		},
	}); err != nil {
		log.Error(ctx, "failed to psuh repository. error: %v", err)
		return err
	}

	log.Info(ctx, "repository push successfuly")
	return nil
}

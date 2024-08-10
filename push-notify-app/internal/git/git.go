package git

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v63/github"
	"github.com/murasame29/image-registry-push-notify/sample-app/internal/log"
)

var (
	ErrNonFastForwardUpdate = git.ErrNonFastForwardUpdate
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

func (g *GitHub) Branch(ctx context.Context, repo *git.Repository, branch string) error {
	workspace, err := repo.Worktree()
	if err != nil {
		log.Error(ctx, "failed to open worktree. error: %v", err)
		return err
	}

	checkoutOption := &git.CheckoutOptions{
		Branch: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch)),
		Create: true,
	}
	if err := workspace.Checkout(checkoutOption); err != nil {
		log.Error(ctx, "failed to check out. error: %v", err)
		mirrorRemoteBranchRefSpec := fmt.Sprintf("refs/heads/%s:refs/heads/%s", branch, branch)
		if err := fetchOrigin(ctx, repo, mirrorRemoteBranchRefSpec); err != nil {
			return fmt.Errorf("faield to fetch origin. error:%v", err)
		}

		if err := workspace.Checkout(checkoutOption); err != nil {
			return fmt.Errorf("faield to checkout branch. error:%v", err)
		}
	}

	return nil
}
func fetchOrigin(ctx context.Context, repo *git.Repository, refSpecStr string) error {
	remote, err := repo.Remote("origin")
	if err != nil {
		log.Error(ctx, "failed to open worktree. error: %v", err)
		return err
	}

	var refSpecs []config.RefSpec
	if refSpecStr != "" {
		refSpecs = []config.RefSpec{config.RefSpec(refSpecStr)}
	}

	if err = remote.Fetch(&git.FetchOptions{
		RefSpecs: refSpecs,
	}); err != nil {
		if err == git.NoErrAlreadyUpToDate {
			fmt.Print("refs already up to date")
		} else {
			return fmt.Errorf("fetch origin failed: %v", err)
		}
	}

	return nil
}

func (g *GitHub) Clone(ctx context.Context, repository string) (*git.Repository, string, error) {
	repoName := strings.Split(repository, "/")[4]
	dir := fmt.Sprintf("/tmp/%s_%d", repoName, time.Now().Unix())
	repo, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL: repository,
		Auth: &http.BasicAuth{
			Username: g.username,
			Password: g.token,
		},
	})
	if err != nil {
		log.Error(ctx, "failed to clone repository. repository: %s error: %v", repository, err)
		return nil, "", err
	}

	return repo, dir, nil
}

func (g *GitHub) Commit(ctx context.Context, repo *git.Repository, path string, message string) (string, error) {
	workspace, err := repo.Worktree()
	if err != nil {
		log.Error(ctx, "failed to open worktree. error: %v", err)
		return "", err
	}

	log.Info(ctx, "trying git add. path: %s", path)

	if _, err := workspace.Add("."); err != nil {
		log.Error(ctx, "failed to git add. path: %s error: %v", path, err)
		return "", err
	}

	log.Info(ctx, "git add successfuly path:%s", path)
	status, err := workspace.Status()
	if err != nil {
		log.Error(ctx, "failed to git status. error: %v", err)
		return "", err
	}

	log.Info(ctx, status.String())

	log.Info(ctx, "trying git commit -m %s.", message)
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

	log.Info(ctx, "git commit successfuly")
	return commit.String(), nil
}

func (g *GitHub) Push(ctx context.Context, repo *git.Repository) error {
	log.Info(ctx, "trying reposiotry push to origin...")
	o := &git.PushOptions{
		Auth: &http.BasicAuth{
			Username: g.username,
			Password: g.token,
		},
	}
	log.Info(ctx, "push options: %v", o)
	if err := o.Validate(); err != nil {
		log.Error(ctx, "failed to validate push options. error: %v", err)
		return err
	}
	log.Info(ctx, "push options validate successfuly")
	log.Info(ctx, "trying open remote")
	remote, err := repo.Remote(o.RemoteName)
	if err != nil {
		log.Error(ctx, "failed to open remote. error: %v", err)
		return err
	}

	log.Info(ctx, "remote open successfuly")
	log.Info(ctx, "trying push")

	if err := remote.PushContext(ctx, o); err != nil {
		return err
	}

	log.Info(ctx, "repository push successfuly")
	return nil
}

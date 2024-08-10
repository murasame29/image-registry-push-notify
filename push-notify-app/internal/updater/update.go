package updater

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/murasame29/image-registry-push-notify/sample-app/internal/git"
	"github.com/murasame29/image-registry-push-notify/sample-app/internal/log"
	"github.com/murasame29/image-registry-push-notify/sample-app/internal/model"
	"gopkg.in/yaml.v2"
	"sigs.k8s.io/kustomize/api/types"
)

var (
	ErrImageTagNotAllowed = errors.New("image tag not allowed")
	ErrImageTagDeny       = errors.New("image tag deny")
	ErrDuplicatePR        = errors.New("duplicate pr")
)

func Update(ctx context.Context, config *AppConfig, event *model.ECRPushEvent) error {
	return update(ctx, config, event)
}
func update(ctx context.Context, config *AppConfig, event *model.ECRPushEvent) error {
	github, err := git.NewGitHub(config.GitHubApplicationID, config.GitHubAppInstallationID, config.GitHubUsername, config.GitHubAppCrtPath)
	if err != nil {
		return fmt.Errorf("failed to new github. error: %v", err)
	}

	regitryConfig, err := config.parseConfig(event)
	if err != nil {
		return fmt.Errorf("failed to new github. error: %v", err)
	}

	if !regitryConfig.checkAllowTag(event.Detail.ImageTag) {
		log.Warn(ctx, "image tag not allowed. event: %v", event)
		return ErrImageTagNotAllowed
	}

	if !regitryConfig.checkDenyTag(event.Detail.ImageTag) {
		log.Warn(ctx, "image tag deny. event: %v", event)
		return ErrImageTagDeny
	}

	repositoryDir, err := regitryConfig.buildRepositoryName(event)
	if err != nil {
		return fmt.Errorf("repository path failed. error: %v", err)
	}

	repoURI := strings.Split(regitryConfig.GitHubRepository, "/")
	repo, filePath, err := github.Clone(ctx, strings.Join(repoURI[:5], "/"))
	if err != nil {
		return fmt.Errorf("failed to clone repository. error: %v", err)
	}

	defer os.RemoveAll(filePath) // error: no check

	targetDir := filepath.Join(filePath, strings.Join(strings.Split(repositoryDir, "/")[5:], "/"), "kustomization.yaml")

	kustomizationFile, err := os.Open(targetDir)
	if err != nil {
		return fmt.Errorf("failed to oepn kustomization.yaml dir. error: %v", err)
	}

	kustomizationData, err := io.ReadAll(kustomizationFile)
	if err != nil {
		return fmt.Errorf("failed to read file. error: %v", err)
	}

	var kustomization types.Kustomization
	if err := yaml.Unmarshal(kustomizationData, &kustomization); err != nil {
		return fmt.Errorf("failed to unmarshal kustomizatioin.yaml. error: %v", err)
	}

	imageURI := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s", event.Account, event.Region, event.Detail.RepositoryName)
	image := findImage(kustomization.Images, imageURI)

	if image == nil {
		// .imagesがないから作る
		kustomization.Images = append(kustomization.Images, types.Image{
			Name:   imageURI,
			NewTag: event.Detail.ImageTag,
		})
	} else {
		image.NewTag = event.Detail.ImageTag
	}

	newKustomization, err := yaml.Marshal(kustomization)
	if err != nil {
		return fmt.Errorf("failed to marshal kustomization. error: %v", err)
	}

	stat, err := kustomizationFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get stat. error: %v", err)
	}

	if err := github.Branch(ctx, repo, fmt.Sprintf("image_updater_%s_%s_%s", strings.Join(strings.Split(event.Detail.RepositoryName, "/")[1:], "_"), regitryConfig.Env[event.Account], event.Detail.ImageTag)); err != nil {
		return fmt.Errorf("faield to switch branch. error: %v", err)
	}

	if err := os.WriteFile(targetDir, newKustomization, stat.Mode().Perm()); err != nil {
		return fmt.Errorf("failed to write file. error: %v", err)
	}

	if _, err := github.Commit(ctx, repo, targetDir, fmt.Sprintf("[%s][image-committer][%s] イメージの更新 ", regitryConfig.Env[event.Account], event.Detail.RepositoryName)); err != nil {
		return fmt.Errorf("failed to commit. error: %v", err)
	}

	if err := github.Push(context.Background(), repo); err != nil {
		// 既にPRがある場合は無視 実装がきったないのは許容　wrapされてて比較できなかった
		if strings.Contains(err.Error(), git.ErrNonFastForwardUpdate.Error()) {
			log.Warn(ctx, "failed to push. error: %v", err)
			return ErrDuplicatePR
		}
		log.Error(ctx, "failed to push. error: %v", err)
		return fmt.Errorf("failed to push. error: %v", err)
	}
	return nil
}

func findImage(images []types.Image, imageURI string) *types.Image {
	for _, image := range images {
		if image.Name == imageURI {
			return &image
		}
	}
	return nil
}

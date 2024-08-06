package updater

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"sigs.k8s.io/kustomize/api/types"

	"github.com/murasame29/image-registry-push-notify/sample-app/internal/git"
	"github.com/murasame29/image-registry-push-notify/sample-app/internal/model"
)

func Update(ctx context.Context, config *AppConfig, event *model.ECRPushEvent) error {
	git, err := git.NewGitHub(config.GitHubApplicationID, config.GitHubAppInstallationID, config.GitHubUsername, config.GitHubAppCrtPath)
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

	repositoryDir, err := regitryConfig.buildRepositoryName(event)
	if err != nil {
		return fmt.Errorf("repository path failed. error: %v", err)
	}

	repoURI := strings.Split(regitryConfig.GitHubRepository, "/")
	repo, filePath, err := git.Clone(ctx, fmt.Sprintf("https://%s", strings.Join(repoURI[:3], "/")))
	if err != nil {
		return fmt.Errorf("failed to clone repository. error: %v", err)
	}

	defer os.RemoveAll(filePath) // error: no check

	targetDir := filepath.Join(filePath, strings.Join(strings.Split(repositoryDir, "/")[3:], "/"), "kustomization.yaml")

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

	if err := os.WriteFile(targetDir, newKustomization, stat.Mode().Perm()); err != nil {
		return fmt.Errorf("failed to write file. error: %v", err)
	}

	if err := git.Branch(ctx, repo, fmt.Sprintf("test_%d", time.Now().Unix())); err != nil {
		return fmt.Errorf("faield to switch branch. error: %v", err)
	}

	if _, err := git.Commit(ctx, repo, targetDir, "fix: kustomization new tag"); err != nil {
		return fmt.Errorf("failed to commit. error: %v", err)
	}
	return git.Push(ctx, repo)
}

func findImage(images []types.Image, imageURI string) *types.Image {
	for _, image := range images {
		if image.Name == imageURI {
			return &image
		}
	}
	return nil
}

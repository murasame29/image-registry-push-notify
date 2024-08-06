package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/murasame29/image-registry-push-notify/sample-app/cmd/config"
	"github.com/murasame29/image-registry-push-notify/sample-app/internal/log"
	"github.com/murasame29/image-registry-push-notify/sample-app/internal/model"
	"github.com/murasame29/image-registry-push-notify/sample-app/internal/updater"
)

func HandleRequest(w http.ResponseWriter, r *http.Request) {
	ctx := log.IntoContext(r.Context(), log.NewLogger(config.Config.App.LogLevel, os.Stdout))

	var eventBody model.ECRPushEvent

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error(ctx, "failed to read body. error: %v", err)
		w.WriteHeader(400)
		return
	}

	if err := json.Unmarshal(body, &eventBody); err != nil {
		log.Error(ctx, "failed to uunmarshal. error: %v", err)
		w.WriteHeader(400)
		return
	}

	log.Debug(ctx, "event recieved! event: %v", eventBody)
	log.Info(ctx, "event recieved! event-type: %s", eventBody.DetailType)

	log.Debug(ctx, "trying parse config...")
	registryConfigs, err := updater.NewConfigWithFile(config.Config.App.ConfigPath)
	if err != nil {
		log.Error(ctx, "failed to parse config. error: %v", err)
		w.WriteHeader(500)
		return
	}

	log.Debug(ctx, "parse config successfly")
	log.Debug(ctx, "trying update")

	if err := updater.Update(ctx, &updater.AppConfig{
		LogLevel:                config.Config.App.LogLevel,
		GitHubApplicationID:     config.Config.GitHub.ApplicationID,
		GitHubAppInstallationID: config.Config.GitHub.InstallationID,
		GitHubUsername:          config.Config.GitHub.Username,
		GitHubAppCrtPath:        config.Config.GitHub.CrtPath,
		RegistryConfig:          registryConfigs,
	}, &eventBody); err != nil {
		log.Error(ctx, "failed to updater. error: %v", err)
		w.WriteHeader(500)
		return
	}

	log.Debug(ctx, "update successfly")

	w.WriteHeader(200)
	fmt.Fprintf(w, "successfly")
}

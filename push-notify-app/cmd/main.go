package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/murasame29/image-registry-push-notify/sample-app/cmd/config"
	"github.com/murasame29/image-registry-push-notify/sample-app/internal/log"
	"github.com/murasame29/image-registry-push-notify/sample-app/internal/model"
	"github.com/murasame29/image-registry-push-notify/sample-app/internal/queue/aws"
	"github.com/murasame29/image-registry-push-notify/sample-app/internal/updater"
)

func init() {
	if err := config.LoadEnv(); err != nil {
		log.Error(context.TODO(), "failed to load env. error: %v", err)
		os.Exit(1)
	}
}

func main() {
	if err := run(); err != nil {
		log.Error(context.Background(), "failed to run. error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := log.IntoContext(context.Background(), log.NewLogger(config.Config.App.LogLevel, os.Stdout))

	log.Debug(ctx, "trying parse config...")
	registryConfigs, err := updater.NewConfigWithFile(config.Config.App.ConfigPath)
	if err != nil {
		log.Error(ctx, "failed to parse config. error: %v", err)
		return err
	}

	go func() {
		for {
			sqs, err := aws.NewSQS(config.Config.AWS.QueueURI, config.Config.AWS.RoleARN)
			if err != nil {
				log.Error(ctx, "failed to create sqs client. error: %v", err)
				continue
			}

			messages, err := sqs.ReceiveMessage(ctx)
			if err != nil {
				log.Error(ctx, "failed to receive message. error: %v", err)
				continue
			}

			for _, message := range messages {
				go func(message types.Message) {
					now := time.Now()
					log.Debug(ctx, "event recieved! event: %v", message)

					var eventBody *model.ECRPushEvent
					if err := json.Unmarshal([]byte(*message.Body), &eventBody); err != nil {
						log.Error(ctx, "failed to unmarshal event. error: %v", err)
						return
					}

					log.Debug(ctx, "event recieved! event: %v", eventBody)

					if err := updater.Update(ctx, &updater.AppConfig{
						LogLevel:                config.Config.App.LogLevel,
						GitHubAppInstallationID: config.Config.GitHub.InstallationID,
						GitHubApplicationID:     config.Config.GitHub.ApplicationID,
						GitHubUsername:          config.Config.GitHub.Username,
						GitHubAppCrtPath:        config.Config.GitHub.CrtPath,
						RegistryConfig:          registryConfigs,
					}, eventBody); validateUpdateError(err) {
						log.Error(ctx, "failed to update. error: %v", err)
						return
					}

					log.Debug(ctx, "update successfly")
					log.Debug(ctx, "trying delete message")

					if err := sqs.DeleteMessage(ctx, *message.ReceiptHandle); err != nil {
						log.Error(ctx, "failed to delete message. error: %v", err)
						return
					}

					log.Debug(ctx, "delete message successfly duration: %d ms", time.Since(now).Milliseconds())
				}(message)
			}

			time.Sleep(config.Config.App.Interval)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	<-sig

	log.Info(ctx, "shutdown successfly by signal")

	return nil
}

// validateUpdateError　は更新処理でエラーとして返されたエラーがinternalのエラーでないかを検証します
func validateUpdateError(err error) bool {
	return err != nil && !errors.Is(err, updater.ErrDuplicatePR) && !errors.Is(err, updater.ErrImageTagDeny) && !errors.Is(err, updater.ErrImageTagNotAllowed)
}

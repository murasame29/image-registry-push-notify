package aws

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type SQS struct {
	queueURI string
	arn      string

	client *sqs.Client
}

func NewSQS(queueURI, arn string) (*SQS, error) {
	// 検証
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load config. error: %w", err)
	}

	client := sts.NewFromConfig(cfg)
	sqsClient := sqs.NewFromConfig(cfg)

	credsCache := aws.NewCredentialsCache(stscreds.NewWebIdentityRoleProvider(
		client,
		arn,
		stscreds.IdentityTokenFile(os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE")),
	))

	creds, err := credsCache.Retrieve(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credentials. error: %w", err)
	}

	if !creds.HasKeys() {
		return nil, errors.New("failed to retrieve credentials")
	}

	return &SQS{
		queueURI: queueURI,
		arn:      arn,
		client:   sqsClient,
	}, nil
}

func (s *SQS) ReceiveMessage(ctx context.Context) ([]types.Message, error) {
	input := &sqs.ReceiveMessageInput{
		QueueUrl:            &s.queueURI,
		VisibilityTimeout:   20,
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     20,
	}

	output, err := s.client.ReceiveMessage(ctx, input)
	if err != nil {
		return nil, err
	}

	return output.Messages, nil
}

func (s *SQS) DeleteMessage(ctx context.Context, receiptHandle string) error {
	input := &sqs.DeleteMessageInput{
		QueueUrl:      &s.queueURI,
		ReceiptHandle: &receiptHandle,
	}

	_, err := s.client.DeleteMessage(ctx, input)
	if err != nil {
		return err
	}

	return nil
}

package s3

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Client wraps the AWS S3 client
type Client struct {
	s3Client *s3.Client
	bucket   string
}

// NewClient creates a new S3 client
func NewClient(ctx context.Context, region, bucket string, useIAMRole bool, accessKey, secretKey string) (*Client, error) {
	var cfg aws.Config
	var err error

	if useIAMRole {
		// Use IAM role (service account in Kubernetes)
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
		)
	} else {
		// Use static credentials (not recommended for production)
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				accessKey,
				secretKey,
				"",
			)),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &Client{
		s3Client: s3.NewFromConfig(cfg),
		bucket:   bucket,
	}, nil
}

// GetObject retrieves an object from S3
func (c *Client) GetObject(ctx context.Context, key string) ([]byte, error) {
	result, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object %s from bucket %s: %w", key, c.bucket, err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object body: %w", err)
	}

	return data, nil
}

// GetIndexYAML retrieves and returns the index.yaml file from S3
func (c *Client) GetIndexYAML(ctx context.Context, key string) ([]byte, error) {
	return c.GetObject(ctx, key)
}

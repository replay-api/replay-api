package s3

import (
	"context"
	"io"
	"log/slog"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
)

type S3Adapter struct {
	Config     common.S3Config
	S3Endpoint string
	BucketName string
	Session    *session.Session
}

func NewS3Adapter(config common.S3Config) *S3Adapter {
	session := session.Must(session.NewSession(&aws.Config{
		Credentials:      credentials.NewEnvCredentials(),
		S3ForcePathStyle: aws.Bool(true),
		Region:           aws.String(endpoints.UsEast1RegionID),
		Endpoint:         aws.String(config.S3Endpoint),
	}))

	return &S3Adapter{
		S3Endpoint: config.S3Endpoint,
		BucketName: config.Bucket,
		Session:    session,
	}
}

func (adapter *S3Adapter) Put(ctx context.Context, replayFileID uuid.UUID, reader io.ReadSeeker) (string, error) {
	c := s3.New(adapter.Session)

	key := replayFileID.String()

	p := s3.PutObjectInput{
		Bucket: aws.String(adapter.BucketName),
		Key:    aws.String(key),
		ACL:    aws.String("public-read"),
		Body:   reader,
	}

	res, err := c.PutObject(&p)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to upload object to s3", err)
		return "", err
	}

	slog.InfoContext(ctx, "Successfully uploaded object to s3", "Response", res)

	uri := adapter.S3Endpoint + "/" + adapter.BucketName + "/" + key

	return uri, nil
}

func (adapter *S3Adapter) GetByID(ctx context.Context, replayFileID uuid.UUID) (io.Reader, error) {
	s := session.Must(session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials("foo", "var", ""),
		S3ForcePathStyle: aws.Bool(true),
		Region:           aws.String(endpoints.UsEast1RegionID),
		Endpoint:         aws.String(adapter.S3Endpoint),
	}))

	client := s3.New(s, &aws.Config{})

	cmd := s3.GetObjectInput{
		Bucket: aws.String(adapter.BucketName),
		Key:    aws.String(replayFileID.String()),
	}

	res, err := client.GetObject(&cmd)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get object from s3", err)
		return nil, err
	}

	reader := res.Body

	return reader, nil
}

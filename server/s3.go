package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	minChunkSize = 5 * 1024 * 1024
)

var (
	awsAccessKeyID     = os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsBucketRegion    = os.Getenv("AWS_REGION")
	awsBucket          = os.Getenv("BUCKET_NAME")

	s3Client  *s3.S3
	setupSync = &sync.Once{}
)

func setupS3Client() {
	creds := credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, "")
	_, err := creds.Get()
	if err != nil {
		log.Fatalf("bad credentials: %s", err)
	}
	cfg := aws.NewConfig().WithRegion(awsBucketRegion).WithCredentials(creds)
	s3Client = s3.New(session.New(), cfg)
}

type S3Uploader struct {
	Client    *s3.S3
	ObjectKey string

	ch     chan Message
	buffer bytes.Buffer

	uploadID *string

	partN         int64
	uploadedParts []*s3.CompletedPart
}

func NewS3Uploader(objKey string) (*S3Uploader, error) {
	setupSync.Do(setupS3Client)

	key, err := pickObjectName(s3Client, objKey, 0)
	if err != nil {
		return nil, fmt.Errorf("could not pick obj key: %w", err)
	}

	return &S3Uploader{
		Client:        s3Client,
		ObjectKey:     key,
		ch:            make(chan Message, 10),
		buffer:        bytes.Buffer{},
		partN:         1,
		uploadedParts: make([]*s3.CompletedPart, 0),
	}, nil
}

func pickObjectName(client *s3.S3, key string, i int) (string, error) {
	keyToTry := key
	if i != 0 {
		keyToTry = fmt.Sprintf("%s-%d", key, i)
	}
	i++

	_, err := client.HeadObject(&s3.HeadObjectInput{
		Bucket: &awsBucket,
		Key:    &keyToTry,
	})

	if err != nil {
		log.Println(err.Error())
		if !strings.Contains(err.Error(), "Not Found") {
			return "", err
		}

		return key, nil
	}

	return pickObjectName(client, key, i)
}

func (up *S3Uploader) Enqueue(m Message) {
	up.ch <- m
}

func (up *S3Uploader) Start(ctx context.Context, wg *sync.WaitGroup) {
	for {
		select {
		case <-ctx.Done():
			if err := up.flush(true); err != nil {
				log.Printf("Could not flush and close uploader: %s", err)
			}
			wg.Done()
			return
		case m := <-up.ch:
			if err := up.appendToBuffer(m); err != nil {
				log.Printf("Could not flush buffer: %s", err)
			}
			log.Printf("Appending message current buff len: %d", up.buffer.Len())
			if up.buffer.Len() >= minChunkSize {
				log.Println("Start flush")
				if err := up.flush(false); err != nil {
					log.Printf("Could not flush buffer: %s", err)
				}
			}
		}
	}
}

func (up *S3Uploader) appendToBuffer(m Message) error {
	bytes := append(m.Payload, byte('\n'))
	up.buffer.Write(bytes)

	return nil
}

func (up *S3Uploader) flush(close bool) error {
	if up.uploadID == nil {
		if err := up.initiateUpload(); err != nil {
			return fmt.Errorf("Could not initiate upload: %w", err)
		}
	}

	if err := up.uploadPart(); err != nil {
		return fmt.Errorf("could not upload part: %w", err)
	}
	up.buffer.Reset()

	if close {
		if err := up.finishUpload(); err != nil {
			return fmt.Errorf("could not finish upload: %w", err)
		}
	}

	return nil
}

func (up *S3Uploader) initiateUpload() error {
	input := &s3.CreateMultipartUploadInput{
		Bucket:      aws.String(awsBucket),
		Key:         aws.String(up.ObjectKey),
		ContentType: aws.String("application/json"),
	}

	resp, err := up.Client.CreateMultipartUpload(input)
	if err != nil {
		return err
	}

	up.uploadID = resp.UploadId
	return nil
}

func (up *S3Uploader) uploadPart() error {
	partInput := &s3.UploadPartInput{
		Body:          bytes.NewReader(up.buffer.Bytes()),
		Bucket:        &awsBucket,
		Key:           &up.ObjectKey,
		PartNumber:    aws.Int64(up.partN),
		UploadId:      up.uploadID,
		ContentLength: aws.Int64(int64(up.buffer.Len())),
	}
	uploadResult, err := up.Client.UploadPart(partInput)
	if err != nil {
		return err
	}
	up.uploadedParts = append(up.uploadedParts, &s3.CompletedPart{
		ETag:       uploadResult.ETag,
		PartNumber: aws.Int64(up.partN),
	})
	up.partN++

	return nil
}

func (up *S3Uploader) finishUpload() error {
	completeInput := &s3.CompleteMultipartUploadInput{
		Bucket:   &awsBucket,
		Key:      &up.ObjectKey,
		UploadId: up.uploadID,
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: up.uploadedParts,
		},
	}

	_, err := up.Client.CompleteMultipartUpload(completeInput)
	return err
}

package s3

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type AmazonS3Info struct {
	AccessKeyID     string
	SerectAccessKey string
	Region          string
	Bucket          string
	Destination     string
}

func UploadS3(info AmazonS3Info, uploadStream io.Reader) {
	config := aws.NewConfig()
	config.Credentials = credentials.NewStaticCredentials(info.AccessKeyID, info.SerectAccessKey, "")
	sess, err := session.NewSession(config)
	if err != nil {
		panic(fmt.Errorf("failed to open new session ", err).Error())
	}

	uploader := s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
		u.S3 = s3.New(sess, aws.NewConfig().WithRegion(info.Region))
	})

	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(info.Bucket),
		Key:    aws.String(info.Destination),
		Body:   uploadStream,
	})
	if err != nil {
		panic(fmt.Errorf("failed to upload file, %v", err).Error())
	}
	fmt.Printf("file uploaded to, %s\n", result.Location)
}

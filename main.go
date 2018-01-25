package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
)

func main() {
	s3Info, postgresInfo, err := ParseFlags()
	if err != nil {
		panic(errors.Wrap(err, "Can't parse info").Error())
	}
	postgresInfo.TempLocalBackupName = "./147068850.sql"

	out, err := exec.Command("pg_dump", postgresInfo.ToCommandOptions()...).Output()
	if err != nil {
		panic(errors.Wrap(err, "Can't dump postgres database"+string(out)).Error())
	}

	f, err := os.Open(postgresInfo.TempLocalBackupName)
	if err != nil {
		panic(fmt.Errorf("failed to open file %q, %v", postgresInfo.TempLocalBackupName, err).Error())
	}
	defer os.Remove(postgresInfo.TempLocalBackupName)

	UploadS3(s3Info, f)
}

type PostgresInfo struct {
	DatabaseName        string
	Server              string
	Port                string
	Username            string
	TempLocalBackupName string
}

func (p PostgresInfo) ToCommandOptions() []string {
	return []string{
		"-Fc",
		fmt.Sprintf("-d%s", p.DatabaseName),
		fmt.Sprintf("-h%s", p.Server),
		fmt.Sprintf("-p%s", p.Port),
		fmt.Sprintf("-U%s", p.Username),
		fmt.Sprintf("-f%s", p.TempLocalBackupName),
	}
}

func ParseFlags() (AmazonS3Info, PostgresInfo, error) {
	postgresDatabaseName := flag.String("PostgresDatabaseName", "", "Postgres database name")
	postgresServer := flag.String("PostgresServer", "", "Postgres server")
	postgresPort := flag.String("PostgresPort", "", "Postgres port")
	postgresUsername := flag.String("PostgresUsername", "", "Postgres username")

	awsAccessKeyID := flag.String("AWSAccessKeyID", "", "AWS access key ID")
	awsSerectAccessKey := flag.String("AWSSerectAccessKey", "", "AWS serect Key")
	awsRegion := flag.String("AWSRegion", "", "AWS region")
	awsS3Bucket := flag.String("AWSBucket", "", "AWS S3 bucket")
	awsS3Path := flag.String("AWSS3Path", "", "AWS S3 path")
	flag.Parse()

	if postgresDatabaseName == nil || *postgresDatabaseName == "" {
		return AmazonS3Info{}, PostgresInfo{}, errors.New("PostgresDatabaseName argument is required")
	}

	if postgresServer == nil || *postgresServer == "" {
		return AmazonS3Info{}, PostgresInfo{}, errors.New("PostgresServer argument is required")
	}

	if postgresPort == nil || *postgresPort == "" {
		return AmazonS3Info{}, PostgresInfo{}, errors.New("PostgresPort argument is required")
	}

	if postgresUsername == nil || *postgresUsername == "" {
		return AmazonS3Info{}, PostgresInfo{}, errors.New("PostgresUsername argument is required")
	}

	postgresInfo := PostgresInfo{
		DatabaseName: *postgresDatabaseName,
		Server:       *postgresServer,
		Port:         *postgresPort,
		Username:     *postgresUsername,
	}

	if awsAccessKeyID == nil || *awsAccessKeyID == "" {
		return AmazonS3Info{}, PostgresInfo{}, errors.New("AWSAccessKeyID argument is required")
	}
	if awsSerectAccessKey == nil || *awsSerectAccessKey == "" {
		return AmazonS3Info{}, PostgresInfo{}, errors.New("AWSSerectAccessKey argument is required")
	}
	if awsRegion == nil || *awsRegion == "" {
		return AmazonS3Info{}, PostgresInfo{}, errors.New("AWSRegion argument is required")
	}
	if awsS3Bucket == nil || *awsS3Bucket == "" {
		return AmazonS3Info{}, PostgresInfo{}, errors.New("AWSBucket argument is required")
	}
	if awsS3Path == nil || *awsS3Path == "" {
		return AmazonS3Info{}, PostgresInfo{}, errors.New("AWSS3Path argument is required")
	}
	now := time.Now()
	s3Info := AmazonS3Info{
		AccessKeyID:     *awsAccessKeyID,
		SerectAccessKey: *awsSerectAccessKey,
		Region:          *awsRegion,
		Bucket:          *awsS3Bucket,
		Destination:     fmt.Sprintf("%s/%d_%d_%d_%d_%d_%d.db_backup", *awsS3Path, now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second()),
	}
	return s3Info, postgresInfo, nil
}

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

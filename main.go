package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/ledongthuc/postgres_to_amazon_s3/postgres"
	"github.com/ledongthuc/postgres_to_amazon_s3/s3"
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

	s3.UploadS3(s3Info, f)
}

func ParseFlags() (s3.AmazonS3Info, postgres.PostgresInfo, error) {
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
		return s3.AmazonS3Info{}, postgres.PostgresInfo{}, errors.New("PostgresDatabaseName argument is required")
	}

	if postgresServer == nil || *postgresServer == "" {
		return s3.AmazonS3Info{}, postgres.PostgresInfo{}, errors.New("PostgresServer argument is required")
	}

	if postgresPort == nil || *postgresPort == "" {
		return s3.AmazonS3Info{}, postgres.PostgresInfo{}, errors.New("PostgresPort argument is required")
	}

	if postgresUsername == nil || *postgresUsername == "" {
		return s3.AmazonS3Info{}, postgres.PostgresInfo{}, errors.New("PostgresUsername argument is required")
	}

	postgresInfo := postgres.PostgresInfo{
		DatabaseName: *postgresDatabaseName,
		Server:       *postgresServer,
		Port:         *postgresPort,
		Username:     *postgresUsername,
	}

	if awsAccessKeyID == nil || *awsAccessKeyID == "" {
		return s3.AmazonS3Info{}, postgres.PostgresInfo{}, errors.New("AWSAccessKeyID argument is required")
	}
	if awsSerectAccessKey == nil || *awsSerectAccessKey == "" {
		return s3.AmazonS3Info{}, postgres.PostgresInfo{}, errors.New("AWSSerectAccessKey argument is required")
	}
	if awsRegion == nil || *awsRegion == "" {
		return s3.AmazonS3Info{}, postgres.PostgresInfo{}, errors.New("AWSRegion argument is required")
	}
	if awsS3Bucket == nil || *awsS3Bucket == "" {
		return s3.AmazonS3Info{}, postgres.PostgresInfo{}, errors.New("AWSBucket argument is required")
	}
	if awsS3Path == nil || *awsS3Path == "" {
		return s3.AmazonS3Info{}, postgres.PostgresInfo{}, errors.New("AWSS3Path argument is required")
	}
	now := time.Now()
	s3Info := s3.AmazonS3Info{
		AccessKeyID:     *awsAccessKeyID,
		SerectAccessKey: *awsSerectAccessKey,
		Region:          *awsRegion,
		Bucket:          *awsS3Bucket,
		Destination:     fmt.Sprintf("%s/%d_%d_%d_%d_%d_%d.db_backup", *awsS3Path, now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second()),
	}
	return s3Info, postgresInfo, nil
}

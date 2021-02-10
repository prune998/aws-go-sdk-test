package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sts"
)

// Use AWS-Go-SDK to connect to a S3 bucket
// Leverage the AWS SDK to grab the Keys/Secrets/Roles from Env vars
//
// set env S3_BUCKET="your-bucket-name"

func main() {
	// read bucket name from env
	m := make(map[string]string)
	for _, e := range os.Environ() {
		parts := strings.Split(e, "=")
		m[parts[0]] = parts[1]
	}
	bucket := m["S3_BUCKET"]
	fmt.Printf("Using S3 bucket '%s' from env S3_BUCKET\n", bucket)

	// SharedConfigEnable seems to not be useful anymore...
	// sess := session.Must(session.NewSessionWithOptions(session.Options{
	// 	SharedConfigState: session.SharedConfigEnable,
	// }))

	// Open the session to AWS
	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:         aws.String(""),
		DisableSSL:       aws.Bool(false),
		S3ForcePathStyle: aws.Bool(false),
	}))

	// Leverage STS API to see how AWS sees us
	svcSts := sts.New(sess)
	result, err2 := svcSts.GetCallerIdentity(nil)
	if err2 != nil {
		fmt.Printf("Error getting caller identity: %v\n", err2)
	}
	output, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(output))

	// create the S3 session
	svc := s3.New(sess)

	// List S3 Buckets
	bl, err := svc.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	fmt.Println(bl)
	fmt.Println("----------------------------------------------")

	// Get the object myfile from the bucket
	// If the bucket is empty you will have an error
	out, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String("myfile"),
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(out)
	fmt.Println("----------------------------------------------")

	// Open the local file "myfile" and upload it into the bucket
	f, err := os.Open("myfile")
	if err != nil {
		fmt.Printf("failed to open file ./myfile, %v", err)
		os.Exit(1)
	}
	defer f.Close()

	uploader := s3manager.NewUploader(sess)
	input := &s3manager.UploadInput{
		ACL:    aws.String("private"),
		Bucket: aws.String(bucket),
		Key:    aws.String("fromGo"),
		Body:   f,
	}
	_, err = uploader.Upload(input)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println("done uploading")
	fmt.Println("----------------------------------------------")

	// Get the object myfile from the bucket
	out3, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String("fromGo"),
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(out3)
	fmt.Println("----------------------------------------------")
}

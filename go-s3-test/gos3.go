package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
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
// set env S3_OBJECT="your-object-name"

func main() {
	// read bucket name from env
	m := make(map[string]string)
	for _, e := range os.Environ() {
		parts := strings.Split(e, "=")
		m[parts[0]] = parts[1]
	}
	bucket := m["S3_BUCKET"]
	object := m["S3_OBJECT"]
	data := m["S3_DATA_FILE"]
	role := m["S3_ROLE"]
	fmt.Printf("Using S3 bucket '%s/%s' from env S3_BUCKET/S3_OBJECT for %s S3_ROLE \n", bucket, object, role)

	// SharedConfigEnable seems to not be useful anymore...
	// sess := session.Must(session.NewSessionWithOptions(session.Options{
	// 	SharedConfigState: session.SharedConfigEnable,
	// }))

	// Open the session to AWS
	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:         aws.String(""),
		DisableSSL:       aws.Bool(false),
		S3ForcePathStyle: aws.Bool(false),
		Region:           aws.String("us-east-1"),
	}))

	// Leverage STS API to see how AWS sees us
	svcSts := sts.New(sess)
	result, err2 := svcSts.GetCallerIdentity(nil)
	if err2 != nil {
		fmt.Printf("Error getting caller identity: %v\n", err2)
	}
	output, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(output))

	// manually assume role
	// Create a STS client
	// roleSvc := sts.New(sess)

	// roleToAssumeArn := role
	// sessionName := "test_session"
	// roleResult, err := roleSvc.AssumeRole(&sts.AssumeRoleInput{
	// 	RoleArn:         &roleToAssumeArn,
	// 	RoleSessionName: &sessionName,
	// })
	// if err != nil {
	// 	fmt.Println("AssumeRole Error", err)
	// }

	// fmt.Println(roleResult.AssumedRoleUser)

	// result2, err2 := roleSvc.GetCallerIdentity(nil)
	// if err2 != nil {
	// 	fmt.Printf("Error getting caller identity: %v\n", err2)
	// }
	// output2, _ := json.MarshalIndent(result2, "", "  ")
	// fmt.Println(string(output2))

	// create the S3 session
	svc := s3.New(sess, &aws.Config{
		Region:                        aws.String("us-east-1"),
		CredentialsChainVerboseErrors: aws.Bool(true),
		LogLevel:                      aws.LogLevel(5),
		// Credentials:                   svcSts.Config.Credentials,
	})

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
	}

	fmt.Println(bl)

	fmt.Println("------- start list object in bucket ----------")

	startAfter := object
	if objectID, err := strconv.Atoi(object); err == nil {
		startAfter = strconv.Itoa(objectID - 1)
	}

	fmt.Printf("Using Object ID %s\n", startAfter)

	objectList, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:     aws.String(bucket),
		MaxKeys:    aws.Int64(1),
		FetchOwner: aws.Bool(true),
		StartAfter: aws.String(startAfter),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				fmt.Println(s3.ErrCodeNoSuchBucket, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
	}
	fmt.Println(objectList)

	fmt.Println("------- start get object ----------")
	// Get the object from the bucket
	// If the bucket is empty you will have an error
	out, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	defer out.Body.Close()

	fmt.Println(out)

	buf := new(bytes.Buffer)
	buf.ReadFrom(out.Body)
	newStr := buf.String()

	fmt.Println(newStr)

	if data == "" {
		return
	}

	fmt.Println("------- start push object ----------")
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

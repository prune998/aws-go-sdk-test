package main

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sts"
)

// This example prog just look up instances by tags
// using the AWS SDK v2
// see https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/ec2-example-manage-instances.html
func main() {
	// name of the instance to filter
	name := "prune-feature-cluster-us-east-1"

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Leverage STS API to see how AWS sees us
	svcSts := sts.New(sess)
	result, err2 := svcSts.GetCallerIdentity(nil)
	if err2 != nil {
		fmt.Printf("Error getting caller identity: %v\n", err2)
	}
	output, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(output))

	// Create EC2 service client
	svc := ec2.New(sess)

	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:eks:cluster-name"),
				Values: []*string{
					aws.String(name),
				},
			},
		},
	}

	res, err := svc.DescribeInstances(params)
	if err != nil {
		fmt.Println("error describing EC2", err)
	}

	// fmt.Println(res.Reservations)
	for _, r := range res.Reservations {
		for _, i := range r.Instances {
			var nt string
			for _, t := range i.Tags {
				if *t.Key == "Name" {
					nt = *t.Value
					break
				}
			}
			fmt.Println(nt, *i.InstanceId, *i.State.Name)
		}
	}
}

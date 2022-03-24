package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2_types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// This example prog just look up instances by tags
// using the AWS SDK v2
// see https://aws.github.io/aws-sdk-go-v2/docs/code-examples/ec2/describeinstances/
func main() {
	// name of the instance to filter
	name := "prune-feature-cluster-us-east-1"

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
	)
	if err != nil {
		fmt.Printf("Error calling AWS\n", err)
		os.Exit(1)
	}

	// Leverage STS API to see how AWS sees us
	svcSts := sts.NewFromConfig(cfg)
	result, err2 := svcSts.GetCallerIdentity(context.TODO(), nil)
	if err2 != nil {
		fmt.Printf("Error getting caller identity: %v\n", err2)
		os.Exit(1)
	}
	output, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(output))

	// Create EC2 service client
	svc := ec2.NewFromConfig(cfg)

	params := &ec2.DescribeInstancesInput{
		Filters: []ec2_types.Filter{
			{
				Name:   aws.String("tag:eks:cluster-name"),
				Values: []string{name},
			},
		},
	}

	res, err := svc.DescribeInstances(context.TODO(), params)
	if err != nil {
		fmt.Println("error describing EC2", err)
	}

	// fmt.Println(res.Reservations)
	for c, r := range res.Reservations {
		for _, i := range r.Instances {
			var nt string
			for _, t := range i.Tags {
				if *t.Key == "Name" {
					nt = *t.Value
					break
				}
			}
			fmt.Println(c, nt, *i.InstanceId, *&i.State.Name)
		}
	}
}

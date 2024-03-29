package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

	// Search all ENIs that belongs to selected instances
	NetworkInterfaceIds := []string{}
	paginator := ec2.NewDescribeInstancesPaginator(svc, params, func(o *ec2.DescribeInstancesPaginatorOptions) {
		o.Limit = 5
	})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(context.TODO())
		if err != nil {
			log.Printf("error: %v", err)
			return
		}
		for c, r := range output.Reservations {
			for _, i := range r.Instances {
				var nt string
				for _, t := range i.Tags {
					if *t.Key == "Name" {
						nt = *t.Value
						break
					}
				}
				fmt.Println(c, nt, *i.InstanceId, i.State.Name, *i.SubnetId)
				for ic, ifs := range i.NetworkInterfaces {
					fmt.Println(ic,
						"'"+*ifs.Description+"'",
						*ifs.SubnetId, ifs.InterfaceType,
						*ifs.PrivateIpAddress,
						aws.ToString(ifs.MacAddress),
						aws.ToString(ifs.NetworkInterfaceId),
						aws.ToInt32(ifs.Attachment.DeviceIndex),
						aws.ToString(ifs.NetworkInterfaceId),
					)
					NetworkInterfaceIds = append(NetworkInterfaceIds, aws.ToString(ifs.NetworkInterfaceId))
				}
			}
		}

	}

	// Build a filter to seach ENIS based on selected ENIS
	// we need to do that as we want `ec2_types.NetworkInterface` and not `ec2_types.InstanceNetworkInterface`
	input := &ec2.DescribeNetworkInterfacesInput{}
	if len(NetworkInterfaceIds) > 0 {
		input.NetworkInterfaceIds = NetworkInterfaceIds
	}

	// Search for the ENIs data
	var ENIs []ec2_types.NetworkInterface
	ENIPaginator := ec2.NewDescribeNetworkInterfacesPaginator(svc, input)
	for ENIPaginator.HasMorePages() {
		output, err := ENIPaginator.NextPage(context.TODO())
		if err != nil {
			log.Printf("error: %v", err)
			return
		}

		ENIs = append(ENIs, output.NetworkInterfaces...)
	}

	fmt.Println("printing ENIs")
	fmt.Println("InstanceId | NetworkInterfaceId | SubnetId | DeviceIndex")
	for _, iface := range ENIs {
		fmt.Println(
			"'"+*iface.Description+"'",
			aws.ToString(iface.SubnetId),
			iface.InterfaceType,
			aws.ToString(iface.PrivateIpAddress),
			aws.ToInt32(iface.Attachment.DeviceIndex),
			aws.ToString(iface.Attachment.InstanceId),
			aws.ToInt32(iface.Attachment.DeviceIndex),
		)
	}
}

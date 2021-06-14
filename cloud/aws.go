package cloud

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

// GetEC2Service initializes and returns an EC2 session
func GetEC2Service(awsRegion string) (ec2iface.EC2API, error) {
	awsSession, err := session.NewSession(&aws.Config{Region: aws.String(awsRegion)})
	if err != nil {
		return nil, err
	}
	return ec2.New(awsSession), nil
}

// DescribeEBSVolumesByClusterName lists all EBS volumes that belong to a given cluster
// Cluster ownership is determined by whether a resource is tagged with "kubernetes.io/cluster/$CLUSTER_NAME: owned"
func DescribeEBSVolumesByClusterName(svc ec2iface.EC2API, clusterName string) ([]*ec2.Volume, error) {
	input := &ec2.DescribeVolumesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String(fmt.Sprintf("tag:kubernetes.io/cluster/%s", clusterName)),
				Values: aws.StringSlice([]string{"owned"}),
			},
		},
	}
	output, err := svc.DescribeVolumes(input)
	if err != nil {
		return nil, err
	}
	return output.Volumes, nil
}

// TagEC2Resources tags a given list of resources ids with a given list of tags
func TagEC2Resources(svc ec2iface.EC2API, resourceIds []string, tags []*ec2.Tag) error {
	input := &ec2.CreateTagsInput{
		Resources: aws.StringSlice(resourceIds),
		Tags:      tags,
	}
	_, err := svc.CreateTags(input)
	return err
}

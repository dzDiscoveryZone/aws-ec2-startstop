package main

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/ec2iface"
)

type mockedEc2Client struct {
	ec2iface.EC2API
	Resp ec2.DescribeInstancesOutput
}

func (m *mockedEc2Client) DescribeInstancesRequest(in *ec2.DescribeInstancesInput) ec2.DescribeInstancesRequest {
	return ec2.DescribeInstancesRequest{
		Request: &aws.Request{
			Data: &m.Resp,
		},
	}
}
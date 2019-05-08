package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/ec2iface"
	"github.com/pkg/errors"
)

func handler() error {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		log.Fatalf("Error loading credentaisl, error: %v", err)
	}
	cfg.Region = endpoints.UsEast1RegionID

	e := Ec2Client{Client: ec2.New(cfg)}

	tagKey := fmt.Sprintf("tag:%s", os.Getenv("START_STOP_KEY"))
	instances, err := e.getInstanceIds(tagKey, []string{os.Getenv("START_STOP_VALUE")})
	if err != nil {
		return errors.Wrapf(err, "error loading Instance ID's, err: %v", instances)
	}

	var state string
	for _, instance := range instances {
		fmt.Printf("Name: %s, ID: %s, Status Value: %s\n", instance.Name, instance.ID, instance.StateName)
		switch instance.StateName {
		case "running":
			log.Printf("instance %s is in state == %s", instance.Name, instance.ID)
			state = "running"
		case "stopped":
			log.Printf("instance %s is in state == %s", instance.Name, instance.ID)
			state = "stopped"
		}
	}
	log.Printf("state == %s", state)

	if state == "running" {
		err = e.stopInstances(instances)
		if err != nil {
			return errors.Wrapf(err, "unable to stop instances: %v\n", instances)
		}
		log.Printf("successfully stopped instances: %v\n", instances)
	} else if state == "stopped" {
		err = e.startInstances(instances)
		if err != nil {
			return errors.Wrapf(err, "unable to start instances: %v\n", instances)
		}
		log.Printf("successfully started instances: %v\n", instances)
	}

	return nil
}

func main() {
	lambda.Start(handler)
}

// Ec2Client will build your EC2 Client for API Calls. Allows mocking
type Ec2Client struct {
	Client ec2iface.EC2API
}

// Instance is a struct for storing the ID and Names of EC2 instances filtered out by tag
type Instance struct {
	ID        string
	Name      string
	StateName string
}

// getInstanceIds return a slice of Instances matching the tags passed in.
func (e *Ec2Client) getInstanceIds(tagValue string, tags []string) ([]Instance, error) {
	req := e.Client.DescribeInstancesRequest(&ec2.DescribeInstancesInput{
		Filters: []ec2.Filter{
			{
				Name:   aws.String(tagValue),
				Values: tags,
			},
		},
	})
	res, err := req.Send()
	if err != nil {
		return []Instance{}, errors.Wrap(err, "error retrieving instances")
	}

	instances := make([]Instance, 0)
	for _, reservation := range res.Reservations {
		for _, instance := range reservation.Instances {
			for _, tag := range instance.Tags {
				if *tag.Key == "Name" {
					instances = append(instances, Instance{
						ID:        *instance.InstanceId,
						Name:      *tag.Value,
						StateName: string(instance.State.Name),
					})
				}
			}
		}
	}
	return instances, nil
}

// startInstances will start any instances given with the slice of Instance
func (e *Ec2Client) startInstances(instances []Instance) error {
	instanceIds := make([]string, 0, len(instances))
	for _, instance := range instances {
		instanceIds = append(instanceIds, instance.ID)
	}

	req := e.Client.StartInstancesRequest(&ec2.StartInstancesInput{
		InstanceIds: instanceIds,
	})
	_, err := req.Send()
	if err != nil {
		return errors.Wrap(err, "could not start instances")
	}
	return nil
}

// stopInstances will stop any instances given with the slice of Instance
func (e *Ec2Client) stopInstances(instances []Instance) error {
	instanceIds := make([]string, 0, len(instances))
	for _, instance := range instances {
		instanceIds = append(instanceIds, instance.ID)
	}

	req := e.Client.StopInstancesRequest(&ec2.StopInstancesInput{
		InstanceIds: instanceIds,
	})
	_, err := req.Send()
	if err != nil {
		return errors.Wrap(err, "could not stop instances")
	}
	return nil
}

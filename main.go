package main

import (
	"fmt"
	"log"
	"time"

	"github.com/TwiN/aws-eks-auto-tagger/cloud"
	"github.com/TwiN/aws-eks-auto-tagger/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

const (
	MaximumFailedExecutionBeforePanic = 10
)

var (
	executionFailedCounter = 0
)

func main() {
	err := config.Initialize()
	if err != nil {
		panic(err)
	}
	ec2Service, err := cloud.GetEC2Service(config.Get().AwsRegion)
	if err != nil {
		panic(err)
	}
	for {
		start := time.Now()
		if err := run(ec2Service); err != nil {
			log.Printf("Error during execution: %s", err.Error())
			executionFailedCounter++
			if executionFailedCounter > MaximumFailedExecutionBeforePanic {
				panic(fmt.Errorf("execution failed %d times: %v", executionFailedCounter, err))
			}
		} else if executionFailedCounter > 0 {
			log.Printf("Execution was successful after %d failed attempts, resetting counter to 0", executionFailedCounter)
			executionFailedCounter = 0
		}
		log.Printf("Execution took %dms, sleeping for %s", time.Since(start).Milliseconds(), config.Get().ExecutionIntervalBetweenEachRun)
		time.Sleep(config.Get().ExecutionIntervalBetweenEachRun)
	}
}

func run(ec2Service ec2iface.EC2API) error {
	volumes, err := cloud.DescribeEBSVolumesByClusterName(ec2Service, config.Get().ClusterName)
	if err != nil {
		return err
	}
	for _, volume := range volumes {
		var tagsToAdd []*ec2.Tag
		for keyOfTagToAdd, valueOfTagToAdd := range config.Get().Tags {
			foundTag := false
			for _, tag := range volume.Tags {
				if aws.StringValue(tag.Key) == keyOfTagToAdd {
					foundTag = true
					if aws.StringValue(tag.Value) == valueOfTagToAdd {
						// The tag already exists, and already has the right value
						log.Printf("[%s] Volume already has tag %s set to %s, skipping", aws.StringValue(volume.VolumeId), keyOfTagToAdd, valueOfTagToAdd)
						break
					} else {
						if config.Get().OverwriteIfDifferentTagValue {
							tagsToAdd = append(tagsToAdd, &ec2.Tag{
								Key:   aws.String(keyOfTagToAdd),
								Value: aws.String(valueOfTagToAdd),
							})
							log.Printf("[%s] Queuing update for tag %s from %s to %s because OverwriteIfDifferentTagValue is set to true", aws.StringValue(volume.VolumeId), keyOfTagToAdd, aws.StringValue(tag.Value), valueOfTagToAdd)
						} else {
							log.Printf("[%s] Not queuing update for tag %s from %s to %s because OverwriteIfDifferentTagValue is set to false", aws.StringValue(volume.VolumeId), keyOfTagToAdd, aws.StringValue(tag.Value), valueOfTagToAdd)
						}
					}
				}
			}
			if !foundTag {
				log.Printf("[%s] Volume doesn't have tag %s, queuing creation of tag %s with value %s", aws.StringValue(volume.VolumeId), keyOfTagToAdd, keyOfTagToAdd, valueOfTagToAdd)
				tagsToAdd = append(tagsToAdd, &ec2.Tag{
					Key:   aws.String(keyOfTagToAdd),
					Value: aws.String(valueOfTagToAdd),
				})
			}
			if len(tagsToAdd) != 0 {
				if config.Get().EbsTaggingEnabled {
					err := cloud.TagEC2Resources(ec2Service, []string{aws.StringValue(volume.VolumeId)}, tagsToAdd)
					if err != nil {
						return err
					}
					// Some may say this is abusing the sleep function, others may say there's no reason to murder
					// an innocent API
					time.Sleep(500 * time.Millisecond)
				} else {
					log.Printf("[%s] Not executing because EbsTaggingEnabled is set to false", aws.StringValue(volume.VolumeId))
				}
			}
		}
	}
	return nil
}

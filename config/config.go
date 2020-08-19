package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

const (
	EnvTagPrefix                    = "TAG_"
	EnvClusterName                  = "CLUSTER_NAME"
	EnvAwsRegion                    = "AWS_REGION"
	EnvEbsTaggingEnabled            = "EBS_TAGGING_ENABLED"
	EnvOverwriteIfDifferentTagValue = "OVERWRITE_IF_DIFFERENT_TAG_VALUE"
	EnvExecutionIntervalInMinutes   = "EXECUTION_INTERVAL_IN_MINUTES"
)

var cfg *config

type config struct {
	// Tags to add to the AWS resources belonging to the cluster.
	// Currently, the AWS resources are limited to: EBS volumes
	Tags map[string]string

	// ClusterName is the name of the EKS cluster.
	// Used to search for EBS volumes that belong to the cluster
	// (i.e. by looking for the tag `kubernetes.io/cluster/$CLUSTER_NAME: owned`)
	ClusterName string

	// AwsRegion is the name of the region the cluster is in (e.g. us-west-2)
	AwsRegion string

	// IntervalBetweenEachRun is the time to wait between each run
	ExecutionIntervalBetweenEachRun time.Duration

	// EbsTaggingEnabled whether to enable tagging for EBS
	EbsTaggingEnabled bool

	// OverwriteIfDifferentTagValue Whether to overwrite the tag if it already exists, but with a different value
	OverwriteIfDifferentTagValue bool
}

// Initialize is used to initialize the application's configuration
func Initialize() error {
	cfg = &config{
		Tags:                            make(map[string]string),
		EbsTaggingEnabled:               true,
		ExecutionIntervalBetweenEachRun: time.Minute * 10,
	}
	for _, keyValue := range os.Environ() {
		parts := strings.Split(keyValue, "=")
		if len(parts) < 2 {
			return errors.New("expected environment to be in the format key=value, but wasn't")
		}
		key := parts[0]
		value := strings.TrimPrefix(keyValue, fmt.Sprintf("%s=", key))
		switch key {
		case EnvClusterName:
			cfg.ClusterName = value
			log.Printf("[config][Initialize] Setting cluster name to %s", value)
		case EnvAwsRegion:
			cfg.AwsRegion = value
			log.Printf("[config][Initialize] Setting AWS region name to %s", value)
		case EnvEbsTaggingEnabled:
			cfg.EbsTaggingEnabled = value == "true"
			log.Printf("[config][Initialize] EBS tagging is set to %s", value)
		case EnvOverwriteIfDifferentTagValue:
			cfg.OverwriteIfDifferentTagValue = value == "true"
			log.Printf("[config][Initialize] Overwrite on different tag value is set to %s", value)
		case EnvExecutionIntervalInMinutes:
			var err error
			cfg.ExecutionIntervalBetweenEachRun, err = time.ParseDuration(fmt.Sprintf("%sm", value))
			if err != nil {
				return fmt.Errorf("invalid execution interval value: %v", err.Error())
			}
			log.Printf("[config][Initialize] Execution interval is set to %s", cfg.ExecutionIntervalBetweenEachRun.String())
		default:
			if strings.HasPrefix(key, EnvTagPrefix) {
				tagName := strings.TrimPrefix(key, EnvTagPrefix)
				cfg.Tags[tagName] = value
				log.Printf("[config][Initialize] Registered tag '%s' with value '%s'", tagName, value)
			}
		}
	}
	if len(cfg.ClusterName) == 0 {
		return errors.New("cluster name must not be empty")
	}
	if len(cfg.AwsRegion) == 0 {
		return errors.New("aws region name must not be empty")
	}
	return nil
}

func Get() *config {
	if cfg == nil {
		panic("config has not been initialized")
	}
	return cfg
}

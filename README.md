# aws-eks-auto-tagger

Automatically tags all EBS volumes that are tagged with 

```
kubernetes.io/cluster/$CLUSTER_NAME: owned
```

There's no way to automatically add one or multiple tags to an EBS that was created by a PersistentVolume.
By leveraging the aforementioned default tag, this application is able to iterate over all EBS volumes that are owned by the cluster and add one or multiple tags to each EBS volume.


## Environment variables

| Key                                | Description | Default value |
| ---------------------------------- | ----------- | ------------- |
| `TAG_<tag_key>`                    | Tag to add to the resources. The tag name will be the name of the variable stripped of the `TAG_` prefix and the tag value will be the value assigned to the variable | N/A |
| `CLUSTER_NAME`                     | Name of the EKS cluster. Used to search for EBS volumes that belong to the cluster (i.e. by looking for the tag `kubernetes.io/cluster/$CLUSTER_NAME: owned`) | `""` required |
| `AWS_REGION`                       | Name of AWS region. | `""` required |
| `EBS_TAGGING_ENABLED`              | Whether to automatically tag EBS volumes or not | `true` |
| `OVERWRITE_IF_DIFFERENT_TAG_VALUE` | Whether to overwrite the tag if it already exists, but with a different value | `false` |
| `EXECUTION_INTERVAL_IN_MINUTES`    | Time to wait between each run in minutes | `10` |


## Permissions

To function properly, this application requires the following permissions on AWS:
- ec2:CreateTags
- ec2:DescribeVolumes


## Developing

Make sure to set the required environment variables, and consider whether you want
the EBS Volumes to be tagged or not. 
If you just want to test it, you can set `EBS_TAGGING_ENABLED`, and the creation of the tags will be skipped.

Your local aws credentials must also be valid (i.e. you can use `awscli`)

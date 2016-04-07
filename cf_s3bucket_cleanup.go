package main

import (
	"flag"
	"strings"
	"time"

	"github.com/allanliu/easylogger"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

var (
	awsRegion = flag.String(
		"aws-region",
		"us-east-1",
		"AWS region",
	)
	bucketFilter = flag.String(
		"bucket-filter",
		"exhibitors3bucket",
		"Search critieria for buckets that fall under CF",
	)
)

type cfS3BucketCleanup struct {
	cfSVC        cloudformationiface.CloudFormationAPI
	s3SVC        s3iface.S3API
	stacks       []*cloudformation.StackSummary
	bucketFilter string
}

func getSessionConfigs() (*session.Session, *aws.Config) {
	return session.New(), &aws.Config{Region: awsRegion}
}

func (c *cfS3BucketCleanup) getAllCfStackNames() {
	params := &cloudformation.ListStacksInput{
		StackStatusFilter: []*string{
			aws.String(cloudformation.StackStatusCreateComplete),
		},
	}
	resp, err := c.cfSVC.ListStacks(params)
	easylogger.LogFatal(err)
	c.stacks = resp.StackSummaries
}

func checkStackBucketbyName(bucketName string, stackName string) bool {
	return strings.Contains(bucketName, stackName)
}

func isCloudformationBucket(bucketName string, bucketFilter string) bool {
	return strings.Contains(bucketName, bucketFilter)
}

func checkStackBucketbyDate(bucketDate time.Time, stackDate time.Time) bool {
	diff := stackDate.Sub(bucketDate).Seconds()
	return diff < 60 && diff > -60
}

func (c *cfS3BucketCleanup) isBucketDeletable(bucket *s3.Bucket) bool {
	for _, stack := range c.stacks {
		if checkStackBucketbyName(*bucket.Name, *stack.StackName) &&
			checkStackBucketbyDate(
				*bucket.CreationDate,
				*stack.CreationTime,
			) {
			return false
		}
	}
	return true
}

func (c *cfS3BucketCleanup) getBucketContents(bucket *s3.Bucket) []*s3.Object {
	resp, err := c.s3SVC.ListObjects(
		&s3.ListObjectsInput{
			Bucket: bucket.Name,
		},
	)
	easylogger.LogFatal(err)
	return resp.Contents
}

func getObjectIDStruct(objects []*s3.Object) []*s3.ObjectIdentifier {
	var result []*s3.ObjectIdentifier
	for _, object := range objects {
		result = append(
			result,
			&s3.ObjectIdentifier{
				Key: object.Key,
			},
		)
	}
	return result
}

func isBucketEmpty(objects []*s3.Object) bool {
	return len(objects) <= 0
}

func (c *cfS3BucketCleanup) emptyBucket(
	bucket *s3.Bucket,
	objects []*s3.Object,
) []*s3.Error {
	resp, err := c.s3SVC.DeleteObjects(
		&s3.DeleteObjectsInput{
			Bucket: bucket.Name,
			Delete: &s3.Delete{
				Objects: getObjectIDStruct(objects),
			},
		},
	)
	easylogger.LogFatal(err)
	if resp.Errors == nil {
		return []*s3.Error{}
	}
	return resp.Errors
}

func (c *cfS3BucketCleanup) removeUnusedCFBuckets() []*s3.Error {
	var (
		errors  = []*s3.Error{}
		objects []*s3.Object
	)
	resp, err := c.s3SVC.ListBuckets(&s3.ListBucketsInput{})
	easylogger.LogFatal(err)
	for _, bucket := range resp.Buckets {
		if isCloudformationBucket(*bucket.Name, c.bucketFilter) &&
			c.isBucketDeletable(bucket) {
			easylogger.Log("This bucket is to be deleted: ", *bucket.Name)
			objects = c.getBucketContents(bucket)
			if !isBucketEmpty(objects) {
				errs := c.emptyBucket(bucket, objects)
				if len(errs) > 0 {
					errors = append(errors, errs...)
					continue
				}
			}
			_, err := c.s3SVC.DeleteBucket(
				&s3.DeleteBucketInput{
					Bucket: bucket.Name,
				},
			)
			easylogger.LogFatal(err)

		}
	}
	return errors
}

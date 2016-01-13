package main

import (
	"flag"
	"fmt"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/s3"
)

func init() {
	flag.Parse()
}

func main() {
	svc := &cfS3BucketCleanup{
		cfSVC:        cloudformation.New(getSessionConfigs()),
		s3SVC:        s3.New(getSessionConfigs()),
		bucketFilter: *bucketFilter,
	}
	svc.getAllCfStackNames()
	errs := svc.removeUnusedCFBuckets()
	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Println("Error: ", err.Message)
		}
	}
}

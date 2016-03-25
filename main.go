package main

import (
	"flag"
	"os"

	"github.com/allanliu/easylogger"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/s3"
)

func init() {
	flag.Parse()
	easylogger.InitializeLog()
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
			easylogger.Log("Error: ", err.Message)
		}
		os.Exit(1)
	}
}

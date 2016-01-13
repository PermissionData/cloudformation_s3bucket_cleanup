# cloudformation_s3bucket_cleanup
[![Build Status](https://travis-ci.org/PermissionData/cloudformation_s3bucket_cleanup.svg?branch=master)](https://travis-ci.org/PermissionData/cloudformation_s3bucket_cleanup)
cloudformation_s3bucket_cleanup is tool used for cleaning up retired S3 buckets used by AWS to provision CloudFormation stacks


## Usage
**compile for your platform using go build**
```bash
// Example for 64bit Ubuntu
$ cd ~/cloudformation_s3_bucket
$ GOOS=linux GOARCH=amd64 go build .
```
**set up aws credentials in ~/.aws/credentials**
```bash
$ cat ~/.aws/credentials
[default]
aws_access_key_id = <aws_access_key_id>
aws_secret_access_key = <aws_secret_access_key>
```
**run tool for backing up**
```bash
$ ./cloudformation_s3bucket_cleanup --aws-region us-east-1 --bucketfilter exhibitors3bucket
```
The default value for aws-region is us-east-1; the default value for bucketfilter is exhibitors3bucket.  If you want these exact parameters you don't need to use the CLI args.

The bucketfilter is the substring that gets appended on every s3bucket created by AWS when provisioning a CloudFormation stack.

The tool only deletes buckets for stacks that are not in the CREATE_COMPLETE state

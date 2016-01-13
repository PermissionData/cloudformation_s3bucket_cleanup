package main

import (
	"testing"
	"time"

	"github.com/PermissionData/cloudformation_s3bucket_cleanup/mock_cloudformationiface"
	"github.com/PermissionData/cloudformation_s3bucket_cleanup/mock_s3iface"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/golang/mock/gomock"
)

type listS3Objects [][]*s3.Object

func getTimeSecondsBeforeNow(s int64) *time.Time {
	t := time.Now().Add(-time.Duration(s) * time.Second)
	return &t
}

func getMocks(t *testing.T) (
	*mock_cloudformationiface.MockCloudFormationAPI,
	*mock_s3iface.MockS3API,
	*gomock.Controller,
) {
	ctrl := gomock.NewController(t)

	mockCloudformationiface := mock_cloudformationiface.NewMockCloudFormationAPI(
		ctrl,
	)
	mockS3Iface := mock_s3iface.NewMockS3API(ctrl)
	return mockCloudformationiface, mockS3Iface, ctrl
}

func TestRemoveUnusedCFBuckets(t *testing.T) {
	mockCloudformationiface, mockS3Iface, ctrl := getMocks(t)
	defer ctrl.Finish()

	validatePositiveResults := func(errs []*s3.Error) {
		numErrors := len(errs)
		if numErrors > 0 {
			t.Errorf(
				"Expected 0 number of errors but got %v", numErrors)
		}
	}

	var (
		happyPathTests = struct {
			tests    []*cfS3BucketCleanup
			bucket1  *s3.Bucket
			objects1 []*s3.Object
			bucket2  *s3.Bucket
			objects2 []*s3.Object
		}{
			tests: []*cfS3BucketCleanup{
				&cfS3BucketCleanup{
					s3SVC: mockS3Iface,
					cfSVC: mockCloudformationiface,
					stacks: []*cloudformation.StackSummary{
						&cloudformation.StackSummary{
							StackName: aws.String("testS3"),
							StackStatus: aws.String(
								cloudformation.StackStatusCreateComplete,
							),
							CreationTime: getTimeSecondsBeforeNow(30),
							StackId:      aws.String("somerandomhash123"),
						},
					},
					bucketFilter: "s3BucketTest",
				},
				&cfS3BucketCleanup{
					s3SVC: mockS3Iface,
					cfSVC: mockCloudformationiface,
					stacks: []*cloudformation.StackSummary{
						&cloudformation.StackSummary{
							StackName: aws.String("testS3"),
							StackStatus: aws.String(
								cloudformation.StackStatusCreateComplete,
							),
							CreationTime: getTimeSecondsBeforeNow(59),
							StackId:      aws.String("somerandomhash123"),
						},
					},
					bucketFilter: "s3BucketTest",
				},
				&cfS3BucketCleanup{
					s3SVC: mockS3Iface,
					cfSVC: mockCloudformationiface,
					stacks: []*cloudformation.StackSummary{
						&cloudformation.StackSummary{
							StackName: aws.String("testS3"),
							StackStatus: aws.String(
								cloudformation.StackStatusCreateComplete,
							),
							CreationTime: getTimeSecondsBeforeNow(-49),
							StackId:      aws.String("somerandomhash123"),
						},
					},
					bucketFilter: "s3BucketTest",
				},
			},
			bucket1: &s3.Bucket{
				CreationDate: getTimeSecondsBeforeNow(10),
				Name:         aws.String("testS3Removal1s3BucketTest"),
			},
			objects1: []*s3.Object{},
			bucket2: &s3.Bucket{
				CreationDate: getTimeSecondsBeforeNow(300),
				Name:         aws.String("testS3Removal2s3BucketTest"),
			},
			objects2: []*s3.Object{
				&s3.Object{
					ETag: aws.String("somerandomhashxxyy09"),
					Key:  aws.String("somethingunique123"),
				},
			},
		}
	)
	for _, test := range happyPathTests.tests {
		mockS3Iface.EXPECT().ListObjects(
			&s3.ListObjectsInput{
				Bucket: happyPathTests.bucket2.Name,
			},
		).Return(
			&s3.ListObjectsOutput{Contents: happyPathTests.objects2},
			nil,
		)
		mockS3Iface.EXPECT().ListBuckets(&s3.ListBucketsInput{}).Return(
			&s3.ListBucketsOutput{
				Buckets: []*s3.Bucket{
					happyPathTests.bucket1,
					happyPathTests.bucket2,
				},
				Owner: &s3.Owner{
					DisplayName: aws.String("SomeOwner"),
					ID:          aws.String("abc123"),
				},
			},
			nil,
		)
		mockS3Iface.EXPECT().DeleteObjects(
			&s3.DeleteObjectsInput{
				Bucket: happyPathTests.bucket2.Name,
				Delete: &s3.Delete{
					Objects: []*s3.ObjectIdentifier{
						&s3.ObjectIdentifier{
							Key: happyPathTests.objects2[0].Key,
						},
					},
				},
			},
		).Return(
			&s3.DeleteObjectsOutput{Errors: []*s3.Error{}},
			nil,
		)
		mockS3Iface.EXPECT().DeleteBucket(
			&s3.DeleteBucketInput{Bucket: happyPathTests.bucket2.Name},
		).Return(&s3.DeleteBucketOutput{}, nil)

		validatePositiveResults(test.removeUnusedCFBuckets())
	}
	var emptyBucketErrorsTests = struct {
		tests    []*cfS3BucketCleanup
		bucket1  *s3.Bucket
		objects1 []*s3.Object
		bucket2  *s3.Bucket
		objects2 []*s3.Object
	}{
		tests: []*cfS3BucketCleanup{

			&cfS3BucketCleanup{
				s3SVC: mockS3Iface,
				cfSVC: mockCloudformationiface,
				stacks: []*cloudformation.StackSummary{
					&cloudformation.StackSummary{
						StackName:    aws.String("testS3"),
						StackStatus:  aws.String(cloudformation.StackStatusCreateComplete),
						CreationTime: getTimeSecondsBeforeNow(30),
						StackId:      aws.String("somerandomhash123"),
					},
				},
				bucketFilter: "s3BucketTest",
			},
			&cfS3BucketCleanup{
				s3SVC: mockS3Iface,
				cfSVC: mockCloudformationiface,
				stacks: []*cloudformation.StackSummary{
					&cloudformation.StackSummary{
						StackName:    aws.String("testS3"),
						StackStatus:  aws.String(cloudformation.StackStatusCreateComplete),
						CreationTime: getTimeSecondsBeforeNow(59),
						StackId:      aws.String("somerandomhash321"),
					},
				},
				bucketFilter: "s3BucketTest",
			},
			&cfS3BucketCleanup{
				s3SVC: mockS3Iface,
				cfSVC: mockCloudformationiface,
				stacks: []*cloudformation.StackSummary{
					&cloudformation.StackSummary{
						StackName:    aws.String("testS3"),
						StackStatus:  aws.String(cloudformation.StackStatusCreateComplete),
						CreationTime: getTimeSecondsBeforeNow(-49),
						StackId:      aws.String("somerandomhash321"),
					},
				},
				bucketFilter: "s3BucketTest",
			},
		},
		bucket1: &s3.Bucket{
			CreationDate: getTimeSecondsBeforeNow(150),
			Name:         aws.String("testS3Removal1s3BucketTest"),
		},
		objects1: []*s3.Object{
			&s3.Object{
				ETag: aws.String("somerandomhashaabb10"),
				Key:  aws.String("somethingunique456"),
			},
		},
		bucket2: &s3.Bucket{
			CreationDate: getTimeSecondsBeforeNow(300),
			Name:         aws.String("testS3Removal2s3BucketTest"),
		},
		objects2: []*s3.Object{
			&s3.Object{
				ETag: aws.String("somerandomhashxxyy09"),
				Key:  aws.String("somethingunique123"),
			},
		},
	}
	for _, test := range emptyBucketErrorsTests.tests {
		mockS3Iface.EXPECT().ListBuckets(&s3.ListBucketsInput{}).Return(
			&s3.ListBucketsOutput{
				Buckets: []*s3.Bucket{
					emptyBucketErrorsTests.bucket1,
					emptyBucketErrorsTests.bucket2,
				},
				Owner: &s3.Owner{
					DisplayName: aws.String("SomeOwner"),
					ID:          aws.String("abc123"),
				},
			},
			nil,
		)
		gomock.InOrder(
			mockS3Iface.EXPECT().ListObjects(
				&s3.ListObjectsInput{
					Bucket: emptyBucketErrorsTests.bucket1.Name,
				},
			).Return(
				&s3.ListObjectsOutput{
					Contents: emptyBucketErrorsTests.objects1,
				},
				nil,
			),
			mockS3Iface.EXPECT().ListObjects(
				&s3.ListObjectsInput{
					Bucket: emptyBucketErrorsTests.bucket2.Name,
				},
			).Return(
				&s3.ListObjectsOutput{
					Contents: emptyBucketErrorsTests.objects2,
				},
				nil,
			),
		)
		gomock.InOrder(
			mockS3Iface.EXPECT().DeleteObjects(
				&s3.DeleteObjectsInput{
					Bucket: emptyBucketErrorsTests.bucket1.Name,
					Delete: &s3.Delete{
						Objects: []*s3.ObjectIdentifier{
							&s3.ObjectIdentifier{
								Key: emptyBucketErrorsTests.objects1[0].Key, //aws.String("somethingunique456"),
							},
						},
					},
				},
			).Return(
				&s3.DeleteObjectsOutput{
					Errors: []*s3.Error{
						&s3.Error{
							Code:    aws.String("403"),
							Key:     aws.String("testkey456"),
							Message: aws.String("Action Prohibited"),
						},
					},
				},
				nil,
			),
			mockS3Iface.EXPECT().DeleteObjects(
				&s3.DeleteObjectsInput{
					Bucket: emptyBucketErrorsTests.bucket2.Name,
					Delete: &s3.Delete{
						Objects: []*s3.ObjectIdentifier{
							&s3.ObjectIdentifier{
								Key: emptyBucketErrorsTests.objects2[0].Key,
							},
						},
					},
				},
			).Return(
				&s3.DeleteObjectsOutput{
					Errors: []*s3.Error{
						&s3.Error{
							Code:    aws.String("502"),
							Key:     aws.String("testkey123"),
							Message: aws.String("Bucket delete Failed"),
						},
					},
				},
				nil,
			),
		)

		errs := test.removeUnusedCFBuckets()

		nErrors := len(errs)
		if nErrors != 2 {
			t.Errorf("Expected %d errors to be returned got %v", 2, nErrors)
		}
	}
	var noCFBucketsTests = struct {
		tests   []*cfS3BucketCleanup
		bucket1 *s3.Bucket
		bucket2 *s3.Bucket
	}{
		tests: []*cfS3BucketCleanup{

			&cfS3BucketCleanup{
				s3SVC: mockS3Iface,
				cfSVC: mockCloudformationiface,
				stacks: []*cloudformation.StackSummary{
					&cloudformation.StackSummary{
						StackName:    aws.String("testS3"),
						StackStatus:  aws.String(cloudformation.StackStatusCreateComplete),
						CreationTime: getTimeSecondsBeforeNow(30),
						StackId:      aws.String("somerandomhash123"),
					},
				},
				bucketFilter: "B3SucketTest",
			},
			&cfS3BucketCleanup{
				s3SVC: mockS3Iface,
				cfSVC: mockCloudformationiface,
				stacks: []*cloudformation.StackSummary{
					&cloudformation.StackSummary{
						StackName:    aws.String("testS3"),
						StackStatus:  aws.String(cloudformation.StackStatusCreateComplete),
						CreationTime: getTimeSecondsBeforeNow(59),
						StackId:      aws.String("somerandomhash123"),
					},
				},
				bucketFilter: "bucketS3Test",
			},
			&cfS3BucketCleanup{
				s3SVC: mockS3Iface,
				cfSVC: mockCloudformationiface,
				stacks: []*cloudformation.StackSummary{
					&cloudformation.StackSummary{
						StackName:    aws.String("testS3"),
						StackStatus:  aws.String(cloudformation.StackStatusCreateComplete),
						CreationTime: getTimeSecondsBeforeNow(-49),
						StackId:      aws.String("somerandomhash123"),
					},
				},
				bucketFilter: "T3SucketBest",
			},
		},
		bucket1: &s3.Bucket{
			CreationDate: getTimeSecondsBeforeNow(150),
			Name:         aws.String("testS3Removal1s3BucketTest"),
		},
		bucket2: &s3.Bucket{
			CreationDate: getTimeSecondsBeforeNow(300),
			Name:         aws.String("testS3Removal2s3BucketTest"),
		},
	}
	for _, test := range noCFBucketsTests.tests {
		mockS3Iface.EXPECT().ListBuckets(&s3.ListBucketsInput{}).Return(
			&s3.ListBucketsOutput{
				Buckets: []*s3.Bucket{
					noCFBucketsTests.bucket1,
					noCFBucketsTests.bucket2,
				},
				Owner: &s3.Owner{
					DisplayName: aws.String("SomeOwner"),
					ID:          aws.String("abc123"),
				},
			},
			nil,
		)
		validatePositiveResults(test.removeUnusedCFBuckets())
	}
	var bucketsEmptyTests = struct {
		tests   []*cfS3BucketCleanup
		bucket1 *s3.Bucket
		bucket2 *s3.Bucket
	}{
		tests: []*cfS3BucketCleanup{
			&cfS3BucketCleanup{
				s3SVC: mockS3Iface,
				cfSVC: mockCloudformationiface,
				stacks: []*cloudformation.StackSummary{
					&cloudformation.StackSummary{
						StackName:    aws.String("testS3"),
						StackStatus:  aws.String(cloudformation.StackStatusCreateComplete),
						CreationTime: getTimeSecondsBeforeNow(30),
						StackId:      aws.String("somerandomhash123"),
					},
				},
				bucketFilter: "s3BucketTest",
			},
			&cfS3BucketCleanup{
				s3SVC: mockS3Iface,
				cfSVC: mockCloudformationiface,
				stacks: []*cloudformation.StackSummary{
					&cloudformation.StackSummary{
						StackName:    aws.String("testS3"),
						StackStatus:  aws.String(cloudformation.StackStatusCreateComplete),
						CreationTime: getTimeSecondsBeforeNow(59),
						StackId:      aws.String("somerandomhash321"),
					},
				},
				bucketFilter: "s3BucketTest",
			},
			&cfS3BucketCleanup{
				s3SVC: mockS3Iface,
				cfSVC: mockCloudformationiface,
				stacks: []*cloudformation.StackSummary{
					&cloudformation.StackSummary{
						StackName:    aws.String("testS3"),
						StackStatus:  aws.String(cloudformation.StackStatusCreateComplete),
						CreationTime: getTimeSecondsBeforeNow(-49),
						StackId:      aws.String("somerandomhash123"),
					},
				},
				bucketFilter: "s3BucketTest",
			},
		},
		bucket1: &s3.Bucket{
			CreationDate: getTimeSecondsBeforeNow(150),
			Name:         aws.String("testS3Removal1s3BucketTest"),
		},
		bucket2: &s3.Bucket{
			CreationDate: getTimeSecondsBeforeNow(300),
			Name:         aws.String("testS3Removal2s3BucketTest"),
		},
	}
	for _, test := range bucketsEmptyTests.tests {
		mockS3Iface.EXPECT().ListBuckets(&s3.ListBucketsInput{}).Return(
			&s3.ListBucketsOutput{
				Buckets: []*s3.Bucket{
					bucketsEmptyTests.bucket1,
					bucketsEmptyTests.bucket2,
				},
				Owner: &s3.Owner{
					DisplayName: aws.String("SomeOwner"),
					ID:          aws.String("abc123"),
				},
			},
			nil,
		)
		gomock.InOrder(
			mockS3Iface.EXPECT().ListObjects(
				&s3.ListObjectsInput{
					Bucket: bucketsEmptyTests.bucket1.Name,
				},
			).Return(
				&s3.ListObjectsOutput{Contents: []*s3.Object{}},
				nil,
			),
			mockS3Iface.EXPECT().ListObjects(
				&s3.ListObjectsInput{
					Bucket: bucketsEmptyTests.bucket2.Name,
				},
			).Return(
				&s3.ListObjectsOutput{Contents: []*s3.Object{}},
				nil,
			),
		)
		gomock.InOrder(
			mockS3Iface.EXPECT().DeleteBucket(
				&s3.DeleteBucketInput{
					Bucket: bucketsEmptyTests.bucket1.Name,
				},
			).Return(&s3.DeleteBucketOutput{}, nil),
			mockS3Iface.EXPECT().DeleteBucket(
				&s3.DeleteBucketInput{
					Bucket: bucketsEmptyTests.bucket2.Name,
				},
			).Return(&s3.DeleteBucketOutput{}, nil),
		)
		validatePositiveResults(test.removeUnusedCFBuckets())
	}
}

func TestEmptyBucket(t *testing.T) {
	mockCloudformationiface, mockS3Iface, ctrl := getMocks(t)
	defer ctrl.Finish()

	var happyPathTests = struct {
		tests    []*cfS3BucketCleanup
		bucket1  *s3.Bucket
		objects1 []*s3.Object
	}{
		tests: []*cfS3BucketCleanup{
			&cfS3BucketCleanup{
				s3SVC: mockS3Iface,
				cfSVC: mockCloudformationiface,
				stacks: []*cloudformation.StackSummary{
					&cloudformation.StackSummary{
						StackName:    aws.String("testS3"),
						StackStatus:  aws.String(cloudformation.StackStatusCreateComplete),
						CreationTime: getTimeSecondsBeforeNow(30),
						StackId:      aws.String("somerandomhash123"),
					},
				},
				bucketFilter: "s3BucketTest",
			},
		},
		bucket1: &s3.Bucket{
			Name:         aws.String("testBucket1"),
			CreationDate: getTimeSecondsBeforeNow(10),
		},
		objects1: []*s3.Object{
			&s3.Object{
				Key: aws.String("somerandomkey123"),
			},
			&s3.Object{
				Key: aws.String("somerandomkey456"),
			},
		},
	}
	for _, test := range happyPathTests.tests {
		mockS3Iface.EXPECT().DeleteObjects(
			gomock.Any(),
		).Return(
			&s3.DeleteObjectsOutput{
				Errors: []*s3.Error{},
			},
			nil,
		)
		errs := test.emptyBucket(
			happyPathTests.bucket1,
			happyPathTests.objects1,
		)
		noErrors := len(errs)
		if noErrors > 0 {
			t.Errorf("Expected 0 errors but got %v", noErrors)
		}
	}
	var negativeTests = struct {
		tests    []*cfS3BucketCleanup
		bucket1  *s3.Bucket
		objects1 []*s3.Object
	}{
		tests: []*cfS3BucketCleanup{
			&cfS3BucketCleanup{
				s3SVC: mockS3Iface,
				cfSVC: mockCloudformationiface,
				stacks: []*cloudformation.StackSummary{
					&cloudformation.StackSummary{
						StackName:    aws.String("testS3"),
						StackStatus:  aws.String(cloudformation.StackStatusCreateComplete),
						CreationTime: getTimeSecondsBeforeNow(30),
						StackId:      aws.String("somerandomhash123"),
					},
				},
				bucketFilter: "s3BucketTest",
			},
		},
		bucket1: &s3.Bucket{
			Name:         aws.String("testBucket1"),
			CreationDate: getTimeSecondsBeforeNow(10),
		},
		objects1: []*s3.Object{
			&s3.Object{
				Key: aws.String("somerandomkey123"),
			},
			&s3.Object{
				Key: aws.String("somerandomkey456"),
			},
		},
	}
	for _, test := range negativeTests.tests {
		mockS3Iface.EXPECT().DeleteObjects(
			gomock.Any(),
		).Return(
			&s3.DeleteObjectsOutput{
				Errors: []*s3.Error{
					&s3.Error{
						Code:    aws.String("403"),
						Key:     aws.String("testkey456"),
						Message: aws.String("Action Prohibited"),
					},
				},
			},
			nil,
		)
		errs := test.emptyBucket(
			happyPathTests.bucket1,
			happyPathTests.objects1,
		)
		nErrors := len(errs)
		if nErrors != 1 {
			t.Errorf("Expected length 1 for errors, got %v", nErrors)
		}
	}
}

func TestIsBucketEmpty(t *testing.T) {

	var happyPathTests = struct {
		tests listS3Objects
	}{
		tests: listS3Objects{
			[]*s3.Object{},
		},
	}
	for _, test := range happyPathTests.tests {
		result := isBucketEmpty(test)
		if result != true {
			t.Errorf("Expected result of 'true' but got %v", result)
		}
	}

	var negativeTests = struct {
		tests listS3Objects
	}{
		tests: listS3Objects{
			{
				&s3.Object{
					Key: aws.String("testkey1"),
				},
			},
			{
				&s3.Object{
					Key: aws.String("testkey1"),
				},
				&s3.Object{
					Key: aws.String("testkey2"),
				},
				&s3.Object{
					Key: aws.String("testkey3"),
				},
			},
		},
	}

	for _, test := range negativeTests.tests {
		result := isBucketEmpty(test)
		if result != false {
			t.Errorf("Expected result of 'false', but got %v", result)
		}
	}
}

func TestGetObjectIDStruct(t *testing.T) {

	validateResult := func(test []*s3.Object, result []*s3.ObjectIdentifier) {
		expectedLength := len(test)
		resultLength := len(result)
		if resultLength != expectedLength {
			t.Errorf(
				"Expected a []*ObjectIdentfier length of %v but got %v",
				expectedLength,
				resultLength,
			)
		}
	}
	var happyPathTests = struct {
		tests listS3Objects
	}{
		tests: listS3Objects{
			[]*s3.Object{
				&s3.Object{
					Key: aws.String("testkey1"),
				},
			},
			[]*s3.Object{
				&s3.Object{
					Key: aws.String("testkey1"),
				},
				&s3.Object{
					Key: aws.String("testkey2"),
				},
				&s3.Object{
					Key: aws.String("testkey3"),
				},
			},
		},
	}
	for _, test := range happyPathTests.tests {
		validateResult(test, getObjectIDStruct(test))
	}

	var negativeTests = struct {
		tests listS3Objects
	}{
		tests: listS3Objects{
			{},
		},
	}

	for _, test := range negativeTests.tests {
		validateResult(test, getObjectIDStruct(test))
	}
}

func TestGetBucketContents(t *testing.T) {
	_, mockS3Iface, ctrl := getMocks(t)
	defer ctrl.Finish()

	var happyPathTests = []struct {
		inputBucket       *s3.Bucket
		listObjectsOutput *s3.ListObjectsOutput
	}{
		{
			inputBucket: &s3.Bucket{
				Name: aws.String("testbucket1"),
			},
			listObjectsOutput: &s3.ListObjectsOutput{
				Contents: []*s3.Object{
					&s3.Object{
						Key: aws.String("testkey1"),
					},
					&s3.Object{
						Key: aws.String("testkey2"),
					},
					&s3.Object{
						Key: aws.String("testkey3"),
					},
				},
			},
		},
	}

	for _, test := range happyPathTests {
		csbc := &cfS3BucketCleanup{
			s3SVC: mockS3Iface,
		}
		mockS3Iface.EXPECT().ListObjects(
			&s3.ListObjectsInput{
				Bucket: test.inputBucket.Name,
			},
		).Times(1).Return(test.listObjectsOutput, nil)
		result := csbc.getBucketContents(test.inputBucket)
		expectedLength := len(test.listObjectsOutput.Contents)
		resultLength := len(result)
		if expectedLength != resultLength {
			t.Errorf("Expected length of %v but got %v", expectedLength, resultLength)
		}
	}
}

func TestIsBucketDeletable(t *testing.T) {
	var happyPathTests = []struct {
		csbc   *cfS3BucketCleanup
		bucket *s3.Bucket
	}{
		{
			csbc: &cfS3BucketCleanup{
				stacks: []*cloudformation.StackSummary{
					&cloudformation.StackSummary{
						StackName:    aws.String("teststack1"),
						CreationTime: getTimeSecondsBeforeNow(300),
					},
					&cloudformation.StackSummary{
						StackName:    aws.String("teststack2"),
						CreationTime: getTimeSecondsBeforeNow(1000),
					},
				},
			},
			bucket: &s3.Bucket{
				Name:         aws.String("teststack1-fjahgjfdhgjur993jdfkdj"),
				CreationDate: getTimeSecondsBeforeNow(500),
			},
		},
		{
			csbc: &cfS3BucketCleanup{
				stacks: []*cloudformation.StackSummary{
					&cloudformation.StackSummary{
						StackName:    aws.String("teststack1"),
						CreationTime: getTimeSecondsBeforeNow(1200),
					},
					&cloudformation.StackSummary{
						StackName:    aws.String("teststack2"),
						CreationTime: getTimeSecondsBeforeNow(9900),
					},
				},
			},
			bucket: &s3.Bucket{
				Name:         aws.String("teststack3-fjahfu88fjdhf"),
				CreationDate: getTimeSecondsBeforeNow(9841),
			},
		},
		{
			csbc: &cfS3BucketCleanup{
				stacks: []*cloudformation.StackSummary{
					&cloudformation.StackSummary{
						StackName:    aws.String("teststack1"),
						CreationTime: getTimeSecondsBeforeNow(30),
					},
					&cloudformation.StackSummary{
						StackName:    aws.String("teststack2"),
						CreationTime: getTimeSecondsBeforeNow(1000),
					},
				},
			},
			bucket: &s3.Bucket{
				Name:         aws.String("teststack1-fjahgjfdhgjur993jdfkdj"),
				CreationDate: getTimeSecondsBeforeNow(91),
			},
		},
		{
			csbc: &cfS3BucketCleanup{
				stacks: []*cloudformation.StackSummary{
					&cloudformation.StackSummary{
						StackName:    aws.String("teststack1"),
						CreationTime: getTimeSecondsBeforeNow(30),
					},
					&cloudformation.StackSummary{
						StackName:    aws.String("teststack2"),
						CreationTime: getTimeSecondsBeforeNow(1000),
					},
				},
			},
			bucket: &s3.Bucket{
				Name:         aws.String("teststack1-fjahgjfdhgjur993jdfkdj"),
				CreationDate: getTimeSecondsBeforeNow(-31),
			},
		},
	}
	for _, test := range happyPathTests {
		result := test.csbc.isBucketDeletable(test.bucket)
		if result != true {
			t.Errorf("Expected output of 'true' but got '%v' ", result)
		}
	}
	var negativeTests = []struct {
		csbc   *cfS3BucketCleanup
		bucket *s3.Bucket
	}{
		{
			csbc: &cfS3BucketCleanup{
				stacks: []*cloudformation.StackSummary{
					&cloudformation.StackSummary{
						StackName:    aws.String("teststack1"),
						CreationTime: getTimeSecondsBeforeNow(300),
					},
					&cloudformation.StackSummary{
						StackName:    aws.String("teststack2"),
						CreationTime: getTimeSecondsBeforeNow(1000),
					},
				},
			},
			bucket: &s3.Bucket{
				Name:         aws.String("teststack1-fjahgjfdhgjur993jdfkdj"),
				CreationDate: getTimeSecondsBeforeNow(330),
			},
		},
		{
			csbc: &cfS3BucketCleanup{
				stacks: []*cloudformation.StackSummary{
					&cloudformation.StackSummary{
						StackName:    aws.String("teststack1"),
						CreationTime: getTimeSecondsBeforeNow(1200),
					},
					&cloudformation.StackSummary{
						StackName:    aws.String("teststack2"),
						CreationTime: getTimeSecondsBeforeNow(9900),
					},
				},
			},
			bucket: &s3.Bucket{
				Name:         aws.String("teststack2-fjahfu88fjdhf"),
				CreationDate: getTimeSecondsBeforeNow(9841),
			},
		},
		{
			csbc: &cfS3BucketCleanup{
				stacks: []*cloudformation.StackSummary{
					&cloudformation.StackSummary{
						StackName:    aws.String("teststack1"),
						CreationTime: getTimeSecondsBeforeNow(30),
					},
					&cloudformation.StackSummary{
						StackName:    aws.String("teststack2"),
						CreationTime: getTimeSecondsBeforeNow(1000),
					},
				},
			},
			bucket: &s3.Bucket{
				Name:         aws.String("teststack1-fjahgjfdhgjur993jdfkdj"),
				CreationDate: getTimeSecondsBeforeNow(89),
			},
		},
	}
	for _, test := range negativeTests {
		result := test.csbc.isBucketDeletable(test.bucket)
		if result != false {
			t.Errorf("Expected output of 'false' but got '%v' ", result)
		}
	}
}

func TestCheckStackBucketByDate(t *testing.T) {

	validateResults := func(result bool, expectedResult bool) {
		if result != expectedResult {
			t.Errorf("Expected output of '%v' but got '%v'", expectedResult, result)
		}
	}
	var happyPathTests = []struct {
		bucketDate *time.Time
		stackDate  *time.Time
	}{
		{
			bucketDate: getTimeSecondsBeforeNow(1000),
			stackDate:  getTimeSecondsBeforeNow(1030),
		},
		{
			bucketDate: getTimeSecondsBeforeNow(1000),
			stackDate:  getTimeSecondsBeforeNow(1059),
		},
		{
			bucketDate: getTimeSecondsBeforeNow(1000),
			stackDate:  getTimeSecondsBeforeNow(941),
		},
		{
			bucketDate: getTimeSecondsBeforeNow(1000),
			stackDate:  getTimeSecondsBeforeNow(1000),
		},
	}
	for _, test := range happyPathTests {
		validateResults(
			checkStackBucketbyDate(*test.bucketDate, *test.stackDate),
			true,
		)
	}

	var negativeTests = []struct {
		bucketDate *time.Time
		stackDate  *time.Time
	}{
		{
			bucketDate: getTimeSecondsBeforeNow(1000),
			stackDate:  getTimeSecondsBeforeNow(1061),
		},
		{
			bucketDate: getTimeSecondsBeforeNow(1000),
			stackDate:  getTimeSecondsBeforeNow(940),
		},
		{
			bucketDate: getTimeSecondsBeforeNow(1000),
			stackDate:  getTimeSecondsBeforeNow(200),
		},
		{
			bucketDate: getTimeSecondsBeforeNow(1000),
			stackDate:  getTimeSecondsBeforeNow(10000),
		},
	}
	for _, test := range negativeTests {
		validateResults(
			checkStackBucketbyDate(*test.bucketDate, *test.stackDate),
			false,
		)
	}
}

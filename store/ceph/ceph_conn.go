package ceph

import (
	"gopkg.in/amz.v1/aws"
	"gopkg.in/amz.v1/s3"
)

var cephConn *s3.S3

func GetCephConnection() *s3.S3 {
	if cephConn != nil {
		return cephConn
	}
	// initial
	auth := aws.Auth{
		AccessKey:"",
		SecretKey:"",
	}

	curRegion := aws.Region{
		Name:"default",
		EC2Endpoint:"http://127.0.0.1",
		S3Endpoint:"http://127.0.0.1",
		S3BucketEndpoint:"",
		S3LocationConstraint:false,
		S3LowercaseBucket:false,
		Sign:aws.SignV2,
	}

	// create S3 connection
	return s3.New(auth, curRegion)
}

// 获得指定的bucket对象
func GetCephBucket(bucket string) *s3.Bucket {
	conn := GetCephConnection()
	return conn.Bucket(bucket)
}

func PutObject(bucket string, path string, data []byte) error {
	return GetCephBucket(bucket).Put(path, data, "octet-stream", s3.PublicRead)
}

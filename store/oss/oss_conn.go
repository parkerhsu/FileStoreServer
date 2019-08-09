package oss

import(
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"FileStoreServer/config"

	"log"
)

var ossCli *oss.Client

// 创建client对象
func Client() *oss.Client {
	if ossCli != nil {
		return ossCli
	}

	ossCli, err := oss.New(config.OSSEndpoint, config.OSSAccessKeyID, config.OSSAccessKeySecret)
	if err != nil {
		log.Println(err)
	}
	return ossCli
}

func Bucket() *oss.Bucket {
	cli := Client()
	if cli != nil {
		bucket, err := cli.Bucket(config.OSSBucket)
		if err != nil {
			log.Println(err)
			return nil
		}
		return bucket
	}
	return nil
}

func DownloadURL(objectName string) string {
	signedUrl, err := Bucket().SignURL(objectName, oss.HTTPGet, 3600)
	if err != nil {
		log.Println(err)
		return ""
	}
	return signedUrl
}

func BuildLifecycleRule(bucketName string) {
	// 表示前缀为test的对象(文件)距最后修改时间30天后过期。
	ruleTest1 := oss.BuildLifecycleRuleByDays("rule1", "test/", true, 30)
	rules := []oss.LifecycleRule{ruleTest1}

	Client().SetBucketLifecycle(bucketName, rules)
}
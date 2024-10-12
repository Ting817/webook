package ioc

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ecodeclub/ekit"
	"github.com/google/wire"
	"os"
	"webook/internal/repository/dao/article"
)

// 还需要 DB
var s3ArticleDAOSet = wire.NewSet(InitS3, article.NewOssDAO)

func InitS3() *s3.S3 {
	// 腾讯云中对标 s3 和 OSS 的产品叫做 COS
	cosId, ok := os.LookupEnv("COS_APP_ID")
	if !ok {
		panic("没有找到环境变量 COS_APP_ID ")
	}
	cosKey, ok := os.LookupEnv("COS_APP_SECRET")
	if !ok {
		panic("没有找到环境变量 COS_APP_SECRET")
	}
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(cosId, cosKey, ""),
		Region:      ekit.ToPtr[string]("ap-nanjing"),
		Endpoint:    ekit.ToPtr[string]("https://cos.ap-nanjing.myqcloud.com"),
		// 强制使用 /bucket/key 的形态
		S3ForcePathStyle: ekit.ToPtr[bool](true),
	})
	if err != nil {
		panic(err)
	}
	return s3.New(sess)
}

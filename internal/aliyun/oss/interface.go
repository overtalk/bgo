package ialiyunoss

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"

	"github.com/overtalk/bgo/core"
)

const ModuleName = "internal.aliyun.oss"

type PolicyToken struct {
	AccessKeyId string `json:"accessid"`
	Host        string `json:"host"`
	Expire      int64  `json:"expire"`
	Signature   string `json:"signature"`
	Policy      string `json:"policy"`
	Directory   string `json:"dir"`
	//Callback    string `json:"callback"`
}
type IAliyunOssModule interface {
	core.IModule

	GetClient() *oss.Client
	GetBucket() *oss.Bucket
	PutObjectFromFile(remotePath, localPath string) error
	GetObjectToFile(remotePath, localPath string) error
	SignGetUrl(objectKey string, expiredSec int64) (string, error)
	ListObject(value string) ([]string, error)
	IsExist(value string) (bool, error)
	DelObject(value ...string) ([]string, error)
	GetPolicyToken(uploadDir string, expireTime int64) (*PolicyToken, error)
}

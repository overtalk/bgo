package caliyunoss

import (
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

func (this *CAliyunOssModule) GetBucket() *oss.Bucket { return this.ossBucket }

func (this *CAliyunOssModule) GetClient() *oss.Client { return this.ossClient }

func (this *CAliyunOssModule) PutObjectFromFile(remotePath, localPath string) error {
	return this.ossBucket.PutObjectFromFile(remotePath, localPath)
}

func (this *CAliyunOssModule) GetObjectToFile(remotePath, localPath string) error {
	return this.ossBucket.GetObjectToFile(remotePath, localPath)
}

func (this *CAliyunOssModule) SignGetUrl(objectKey string, expiredSec int64) (string, error) {
	return this.ossBucket.SignURL(objectKey, oss.HTTPGet, expiredSec)
}

func (this *CAliyunOssModule) ListObject(value string) ([]string, error) {
	obj, err := this.ossBucket.ListObjects(oss.Prefix(value))
	if err != nil {
		return nil, err
	}

	var ret []string
	for _, v := range obj.Objects {
		// if size > 0, not dir
		if v.Size != 0 {
			ret = append(ret, strings.Trim(v.Key, value+"/"))
		}
	}

	return ret, nil
}

func (this *CAliyunOssModule) IsExist(value string) (bool, error) {
	return this.ossBucket.IsObjectExist(value)
}

func (this *CAliyunOssModule) DelObject(value ...string) ([]string, error) {
	res, err := this.ossBucket.DeleteObjects(value)
	if err != nil {
		return nil, err
	}

	return res.DeletedObjects, err
}

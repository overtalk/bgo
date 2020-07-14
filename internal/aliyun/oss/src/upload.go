package caliyunoss

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"time"

	"github.com/overtalk/bgo/internal/aliyun/oss"
)

type ConfigStruct struct {
	Expiration string     `json:"expiration"`
	Conditions [][]string `json:"conditions"`
}

//type PolicyToken struct {
//	AccessKeyId string `json:"accessid"`
//	Host        string `json:"host"`
//	Expire      int64  `json:"expire"`
//	Signature   string `json:"signature"`
//	Policy      string `json:"policy"`
//	Directory   string `json:"dir"`
//	//Callback    string `json:"callback"`
//}

//type CallbackParam struct {
//	CallbackUrl      string `json:"callbackUrl"`
//	CallbackBody     string `json:"callbackBody"`
//	CallbackBodyType string `json:"callbackBodyType"`
//}

func (this *CAliyunOssModule) GetPolicyToken(uploadDir string, expireTime int64) (*ialiyunoss.PolicyToken, error) {
	now := time.Now().Unix()
	expireEnd := now + expireTime
	var tokenExpire = time.Unix(expireEnd, 0).Format("2006-01-02T15:04:05Z")

	//create post policy json
	var config ConfigStruct
	config.Expiration = tokenExpire
	var condition []string
	condition = append(condition, "starts-with")
	condition = append(condition, "$key")
	condition = append(condition, uploadDir)
	config.Conditions = append(config.Conditions, condition)

	//calucate signature
	result, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	debyte := base64.StdEncoding.EncodeToString(result)
	h := hmac.New(func() hash.Hash { return sha1.New() }, []byte(this.cfg.Sk))
	if _, err := io.WriteString(h, debyte); err != nil {
		return nil, err
	}

	signedStr := base64.StdEncoding.EncodeToString(h.Sum(nil))

	//var callbackParam CallbackParam
	//callbackParam.CallbackUrl = callbackUrl
	//callbackParam.CallbackBody = "filename=${object}&size=${size}&mimeType=${mimeType}&height=${imageInfo.height}&width=${imageInfo.width}"
	//callbackParam.CallbackBodyType = "application/x-www-form-urlencoded"
	//callbackStr, err := json.Marshal(callbackParam)
	//if err != nil {
	//	fmt.Println("callback json err:", err)
	//}
	//callbackBase64 := base64.StdEncoding.EncodeToString(callbackStr)

	var policyToken ialiyunoss.PolicyToken
	policyToken.AccessKeyId = this.cfg.Ak
	policyToken.Host = fmt.Sprintf("http://%s.%s", this.cfg.Bucket, this.cfg.Endpoint)
	policyToken.Expire = expireEnd
	policyToken.Signature = string(signedStr)
	policyToken.Directory = uploadDir
	policyToken.Policy = string(debyte)
	//policyToken.Callback = string(callbackBase64)
	return &policyToken, nil
}

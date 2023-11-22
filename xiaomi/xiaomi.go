package xiaomi

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	BASEURL     = "https://rest-iap.miglobalpay.com"
	ACKNOWLEDGE = "acknowledge"
	CONSUME     = "consume"
)

type MiGlobalIAP struct {
	AppID       string
	AppKey      string
	AppSecret   string
	PackageName string
	BaseUrl     string
}

type PurchaseResult struct {
	Kind                        string `json:"kind"`                        //购买类型：inapp（一次性商品）
	ProductId                   string `json:"productId"`                   //商品ID
	Quantity                    int    `json:"quantity"`                    //商品购买数量
	OrderId                     string `json:"orderId"`                     //订单ID
	PurchaseToken               string `json:"purchaseToken"`               //购买令牌
	PurchaseTimeMillis          string `json:"purchaseTimeMillis"`          //购买时间戳（从1970年1月1日以来的毫秒数)
	PurchaseState               int    `json:"purchaseState"`               //购买状态：0（已购买）；1（已退款）；2（处理中）
	AcknowledgementState        int    `json:"acknowledgementState"`        //确认状态：0（未确认）；1（已确认）
	ConsumptionState            int    `json:"consumptionState"`            //消耗状态：0（未消耗）；1（已消耗）
	PurchaseType                int    `json:"purchaseType"`                //购买类型。正常购买时不返回该字段；仅在非正常购买时返回：0（许可测试，通过许可测试白名单账号购买）
	DeveloperPayload            string `json:"developerPayload"`            //开发者在确认/消耗购买时指定的额外信息
	ObfuscatedExternalAccountId string `json:"obfuscatedExternalAccountId"` //开发者在SDK发起购买时设置的用户账号信息
	ObfuscatedExternalProfileId string `json:"obfuscatedExternalProfileId"` //开发者在SDK发起购买时设置的用户相关信息
	RegionCode                  string `json:"regionCode"`                  //购买区域码（遵循ISO 3166-1标准，2位大写字母）
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func newResponseError(b []byte) *ResponseError {
	err := new(ResponseError)
	json.Unmarshal(b, err)
	return err
}

func (e *ResponseError) Error() string {
	return fmt.Sprintf("code:%d,message:%s", e.Code, e.Message)
}

func newPurchaseResult(b []byte) *PurchaseResult {
	result := new(PurchaseResult)
	json.Unmarshal(b, result)
	return result
}

func New(id, key, secret, pkg string, url ...string) *MiGlobalIAP {
	baseUrl := BASEURL
	if len(url) == 1 {
		baseUrl = url[0]
	}

	return &MiGlobalIAP{
		AppID:       id,
		AppKey:      key,
		AppSecret:   secret,
		PackageName: pkg,
		BaseUrl:     baseUrl,
	}
}

// PurchaseStatus 查询交易状态
func (mi *MiGlobalIAP) PurchaseStatus(pid, token string) (*PurchaseResult, error) {
	if err := mi.verifyToken(token); err != nil {
		return nil, err
	}

	requestUrl := mi.purchaseUrl(pid, token)
	ias, sign, _ := mi.signature(http.MethodGet, requestUrl)
	headers := mi.buildHeader(ias, sign)
	url := mi.BaseUrl + requestUrl
	code, body, err := mi.request(http.MethodGet, url, headers)

	if err != nil {
		return nil, err
	}

	if code != http.StatusOK {
		return nil, newResponseError(body)
	}

	return newPurchaseResult(body), nil
}

// Acknowledge 确认购买
func (mi *MiGlobalIAP) Acknowledge(pid, token string, payload ...string) (bool, error) {
	if err := mi.verifyToken(token); err != nil {
		return false, err
	}

	requestUrl := mi.acknowledgeUrl(pid, token)
	ias, sign, md5Sign := mi.signature(http.MethodPost, requestUrl, payload...)
	headerMap := mi.buildHeader(ias, sign, md5Sign)
	url := mi.BaseUrl + requestUrl
	code, body, err := mi.request(http.MethodPost, url, headerMap, payload...)

	if err != nil {
		return false, err
	}

	if code != http.StatusOK {
		return false, newResponseError(body)
	}

	return true, nil
}

// Consume 消耗购买
func (mi *MiGlobalIAP) Consume(pid, token string, payload ...string) (bool, error) {
	if err := mi.verifyToken(token); err != nil {
		return false, err
	}

	requestUrl := mi.consumeUrl(pid, token)
	ias, sign, md5Sign := mi.signature(http.MethodPost, requestUrl, payload...)
	headerMap := mi.buildHeader(ias, sign, md5Sign)
	url := mi.BaseUrl + requestUrl
	code, body, err := mi.request(http.MethodPost, url, headerMap, payload...)

	if err != nil {
		return false, err
	}

	if code != http.StatusOK {
		return false, newResponseError(body)
	}

	return true, nil
}

// 简单验证一下token
func (mi *MiGlobalIAP) verifyToken(token string) error {
	if len(token) < 3 {
		return errors.New("token error")
	}

	return nil
}

// 查询购买请求地址
func (mi *MiGlobalIAP) purchaseUrl(pid, token string) string {
	return mi.buildRequestUrl(pid, token)
}

// 确认购买请求地址
func (mi *MiGlobalIAP) acknowledgeUrl(pid, token string) string {
	return mi.buildRequestUrl(pid, token, ACKNOWLEDGE)
}

// 消耗购买请求地址
func (mi *MiGlobalIAP) consumeUrl(pid, token string) string {
	return mi.buildRequestUrl(pid, token, CONSUME)
}

// 处理签名
func (mi *MiGlobalIAP) signature(method, url string, payload ...string) (string, string, string) {
	var (
		body       string
		ias        string
		md5Sign    string
		sha1Sign   string
		text       string
		hasPayload bool
	)

	ias = uuid.New().String()

	// 存在请求体的情况
	if len(payload) >= 1 && payload[0] != "" {
		body = payload[0]
		hasPayload = true
	}

	// 测试用，自定义ias签名
	if len(payload) == 2 && payload[1] != "" {
		ias = payload[1]
	}

	if hasPayload {
		md5Sign = mi.md5(body)
		text = fmt.Sprintf("%s\n%s\napplication/json\n\nx-ias-sign-nonce:%s\n%s", method, md5Sign, ias, url)
		sha1Sign = mi.sha1(text)
		return ias, sha1Sign, md5Sign
	}

	text = fmt.Sprintf("%s\n\napplication/json\n\nx-ias-sign-nonce:%s\n%s", method, ias, url)
	sha1Sign = mi.sha1(text)
	return ias, sha1Sign, ""
}

// MD5签名
func (mi *MiGlobalIAP) md5(text string) string {
	b := md5.Sum([]byte(text))
	return strings.ToUpper(base64.StdEncoding.EncodeToString(b[:]))
}

// HMACSHA1签名
func (mi *MiGlobalIAP) sha1(text string) string {
	mac := hmac.New(sha1.New, []byte(mi.AppSecret))
	mac.Write([]byte(text))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// 构建请求地址
func (mi *MiGlobalIAP) buildRequestUrl(pid, token string, opt ...string) string {
	format := "/%s/developer/v1/applications/%s/purchases/products/%s/tokens/%s"

	if len(opt) > 0 && (opt[0] == ACKNOWLEDGE || opt[0] == CONSUME) {
		format = format + ":" + opt[0]
	}

	region := token[1:3]

	return fmt.Sprintf(format, region, mi.PackageName, pid, token)
}

// 构建请求头
func (mi *MiGlobalIAP) buildHeader(ias, sha1 string, md5 ...string) map[string]string {
	m := make(map[string]string)

	m["appId"] = mi.AppID
	m["timestamp"] = strconv.Itoa(int(time.Now().Unix()))
	m["x-ias-sign-nonce"] = ias
	m["Content-Type"] = "application/json"
	if len(md5) == 1 {
		m["Content-MD5"] = md5[0]
	}
	m["Authorization"] = "ias " + sha1

	return m
}

// 发送查询请求
func (mi *MiGlobalIAP) request(method, url string, header map[string]string, payload ...string) (int, []byte, error) {
	var (
		reader io.Reader
		client http.Client
	)

	if payload != nil && len(payload) == 1 {
		reader = strings.NewReader(payload[0])
	} else {
		reader = nil
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return 0, nil, err
	}

	for key, val := range header {
		req.Header.Set(key, val)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}

	return resp.StatusCode, b, nil
}

// Acknowledgement 是否已确认购买
func (s *PurchaseResult) Acknowledgement() bool {
	return s.AcknowledgementState == 1
}

// Consumption 是否已消耗购买
func (s *PurchaseResult) Consumption() bool {
	return s.ConsumptionState == 1
}

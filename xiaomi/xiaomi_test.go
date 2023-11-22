package xiaomi

import (
	"net/http"
	"testing"
)

const (
	PackageName = "com.iap.test"
	AppID       = "123456"
	AppKey      = "123456"
	AppSecret   = "MTIzNDU2Nzg5MA=="
	Sku         = "game-10"
	Token       = "PID220706155637286289388715727088"
	ContentType = "application/json"
)

func TestPurchase(t *testing.T) {
	testIas := "8420674727e94e7cabff27ac66db5c5d"
	expectedSign := "J5yTx/APWWoOhF4owiRbQaqgang="
	expectedHeaders := map[string]string{
		"appId":            AppID,
		"timestamp":        "1664553600",
		"x-ias-sign-nonce": testIas,
		"Content-Type":     ContentType,
		"Authorization":    "ias " + expectedSign,
	}

	iap := New(AppID, AppKey, AppSecret, PackageName)
	requestUrl := iap.purchaseUrl(Sku, Token)
	ias, sign, _ := iap.signature(http.MethodGet, requestUrl, "", testIas)
	headers := iap.buildHeader(ias, sign)

	if sign != expectedSign {
		t.Errorf("sign error:%s, expected:%s", sign, expectedSign)
	}

	for k, v := range expectedHeaders {
		if str, ok := headers[k]; !ok {
			t.Errorf("header does not contain key: %s", k)
		} else {
			if str != v && k != "timestamp" {
				t.Errorf("header %s error:%s, expected:%s", k, str, v)
			}
		}
	}
}

func TestConsume(t *testing.T) {
	testIas := "4e0f4db3558b4eab94092543988947d9"
	expectedSign := "sc/oyzmvcfItnyqQqJy8l+FAN2Y="
	expectedMd5 := "HPJWVBUIVADK4Q9UJB0MTG=="
	expectedHeaders := map[string]string{
		"appId":            AppID,
		"timestamp":        "1664553605",
		"x-ias-sign-nonce": testIas,
		"Content-Type":     ContentType,
		"Content-MD5":      expectedMd5,
		"Authorization":    "ias " + expectedSign,
	}
	payload := `{"developerPayload":"test"}`

	iap := New(AppID, AppKey, AppSecret, PackageName)
	requestUrl := iap.consumeUrl(Sku, Token)
	ias, sign, contentMd5 := iap.signature(http.MethodPost, requestUrl, payload, testIas)
	headers := iap.buildHeader(ias, sign, contentMd5)

	if sign != expectedSign {
		t.Errorf("sign error:%s, expected:%s", sign, expectedSign)
	}

	if contentMd5 != expectedMd5 {
		t.Errorf("content md5 error:%s, expected:%s", contentMd5, expectedMd5)
	}

	for k, v := range expectedHeaders {
		if str, ok := headers[k]; !ok {
			t.Errorf("header does not contain key: %s", k)
		} else {
			if str != v && k != "timestamp" {
				t.Errorf("header %s error:%s, expected:%s", k, str, v)
			}
		}
	}
}

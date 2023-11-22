# go-mi-global-iap
小米国际应用商店，内购商品的服务端验证库

小米官方文档：https://global.developer.mi.com/document?doc=iapDocument.sdkIntegratedApplication


> ✅ 实现内购商品交易状态查询
> 
> ✅ 实现确认购买
> 
> ✅ 实现消耗购买


> ⚠️ 暂时没有实现订阅查询接口

1. 安装

```shell
go get -u github.com/climberone/go-mi-global-iap
```

2. 示例

```golang
package main

import (
	"fmt"
	"github.com/climberone/go-mi-global-iap/xiaomi"
)

const (
	PackageName = "your package name"
	AppID       = "your app id"
	AppKey      = "your app key"
	AppSecret   = "your app secret"
)

func main() {
	pid := "your product id"
	token := "your purchase token"

	iap := xiaomi.New(AppID, AppKey, AppSecret, PackageName)

	// 查询购买
	if result, err := iap.PurchaseStatus(pid, token); err != nil {
		if resp, ok := err.(*xiaomi.ResponseError); ok {
			// 请求成功，平台返回的错误信息
			// https://global.developer.mi.com/document?doc=iapDocument.resolveBillingResult
			fmt.Printf("%#v", resp)
		} else {
			fmt.Println(err)
		}
	} else {
		// 返回 xiaomi.PurchaseResult
		fmt.Printf("%#v", result)

		// 已确认购买并且已消耗
		if result.Acknowledgement() && result.Consumption() {
			//
			// TODO 处理你的业务逻辑
			//
		}
	}

	// 确认购买
	isAcknowledged, _ := iap.Acknowledge(pid, token)
	fmt.Println(isAcknowledged)

	// 消耗购买
	isConsumed, _ := iap.Consume(pid, token, `{"developerPayload":"test"}`) // 有请请求体的情况
	fmt.Println(isConsumed)

}

```
# smsutils-qcloud

Tencent Cloud (QCloud) SMS driver for [smsutils](https://go.gh.ink/smsutils).

This module implements the `model.Driver` / `model.Client` interfaces of the core
`smsutils` library on top of the Tencent Cloud SMS API (`sms` v20210111).

Driver name: `qcloud`

## Installation

```bash
go get go.gh.ink/smsutils/qcloud/v3
```

## Usage

Blank-import this package so it registers itself, then configure it via the core client.

```go
import (
	"go.gh.ink/smsutils/v3/client"
	"go.gh.ink/smsutils/v3/model"

	_ "go.gh.ink/smsutils/qcloud/v3"
)

clients, err := client.NewClient(model.Config{
	Credentials: model.C{
		"qcloud": {
			"secretID":    "<your-secret-id>",
			"secretKey":   "<your-secret-key>",
			"smsSdkAppID": "<your-sms-sdk-app-id>",
			// Optional overrides:
			// "region":   "ap-guangzhou",
			// "endpoint": "sms.tencentcloudapi.com",
		},
	},
})
if err != nil {
	panic(err)
}

err = clients["qcloud"].SendMessage("+8617601205205", "YourSignName", "1234567", model.Vars{
	{Key: "code", Value: "1234"},
})
```

## Context-aware sending

The existing `client.NewClient` and `SendMessage` APIs remain available. With the
updated local core and this driver, use the same configuration with the new factory:

```go
// Additional imports: "context", "time".
clients, err := client.NewClientCtx(config, model.ClientOptions{Timeout: 5 * time.Second})
if err != nil {
    return err
}
ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
defer cancel()
err = clients["qcloud"].SendMessageCtx(ctx, dest, sender, template, vars)
```

`config` has the credential layout shown above. `ClientOptions.Transport` optionally
injects a concurrency-safe HTTP transport. Context cancellation reaches the HTTP request;
a timeout does not prove the provider did not accept the SMS, so avoid blind retries.
Nil variable entries and malformed provider responses are rejected.

These additions require the updated core source. Existing dependency versions are
intentionally unchanged until release. For local testing, create a Go workspace containing
this module and the sibling `smsutils/v3` core module; see the core repository's
`CONTEXT.md` (local path: `../smsutils/CONTEXT.md`) for the complete six-module
commands and API contracts.

## Credentials

| Key | Constant | Required | Default | Description |
|-----|----------|----------|---------|-------------|
| `secretID` | `SecretID` | Yes | — | Tencent Cloud SecretId |
| `secretKey` | `SecretKey` | Yes | — | Tencent Cloud SecretKey |
| `smsSdkAppID` | `SmsSdkAppID` | Yes (per send) | — | SMS application ID (`SmsSdkAppId`) |
| `region` | `Region` | No | `ap-nanjing` | API region |
| `endpoint` | `Endpoint` | No | `sms.tencentcloudapi.com` | API endpoint |

Missing `secretID` or `secretKey` yields `errors.ErrDriverCredentialInvalid`.

## Behavior

- The destination is normalized via `utils.ProcessNumberForChinese` before sending.
- `template` maps to the Tencent `TemplateId`; `sender` maps to `SignName`.
- `vars` are passed **positionally** as `TemplateParamSet` (in the order supplied), so the
  order of `model.Vars` must match the placeholders defined in your template. The `Key`
  fields are ignored.
- The response's `SendStatusSet` is inspected; any entry whose `Code` is not `Ok` produces
  `errors.ErrDriverSendFailed`, decorated with the provider code, message, request ID and raw
  response.

## Requirements

- Go 1.25.0+
- `go.gh.ink/smsutils/v3`
- Tencent Cloud SDKs: `tencentcloud-sdk-go/.../common`, `tencentcloud-sdk-go/.../sms`

## License

[Apache License 2.0](LICENSE)

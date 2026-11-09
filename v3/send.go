package qcloud

import (
	"context"

	smsv20210111 "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"go.gh.ink/smsutils/v3/errors"
	"go.gh.ink/smsutils/v3/model"
	"go.gh.ink/smsutils/v3/utils"
	"go.gh.ink/toolbox/pointer"
)

// SendMessage retains the original API and delegates to the context-aware path.
func (c Client) SendMessage(dest string, sender string, template string, vars model.Vars) error {
	return c.SendMessageCtx(context.Background(), dest, sender, template, vars)
}

// SendMessageCtx binds cancellation to the actual provider HTTP request.
func (c Client) SendMessageCtx(ctx context.Context, dest, sender, template string, vars model.Vars) (err error) {
	if ctx == nil {
		return errors.ErrMessageInvalid
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	defer func() {
		if ctx.Err() != nil {
			err = ctx.Err()
		}
	}()
	if err := utils.ValidateVars(vars, false); err != nil {
		return err
	}

	// Try to parse number
	dest, _, _, _, err = utils.ProcessNumberForChinese(dest)
	if err != nil {
		return err
	}

	// Preprocess vars
	params := make([]*string, len(vars))
	for k, v := range vars {
		params[k] = &v.Value
	}

	// Construct a request
	request := smsv20210111.NewSendSmsRequest()

	// Set request params
	request.PhoneNumberSet = []*string{&dest}
	request.SmsSdkAppId = &c.SmsSdkAPPID
	request.TemplateId = &template
	request.SignName = &sender
	request.TemplateParamSet = params

	// Send requests
	if c.Client == nil {
		return errors.ErrDriverSendFailed.WithDriverName(Name)
	}
	response, err := c.Client.SendSmsWithContext(ctx, request)
	if err != nil {
		return err
	}
	if response != nil && response.Response != nil && len(response.Response.SendStatusSet) == 1 {
		for _, status := range response.Response.SendStatusSet {
			if status == nil {
				return errors.ErrDriverSendFailed.WithDriverName(Name)
			}
			if pointer.SafeDeref(status.Code) != "Ok" || pointer.SafeDeref(status.PhoneNumber) != dest {
				return errors.ErrDriverSendFailed.
					WithDriverName(Name).
					WithDriverCode(pointer.SafeDeref(status.Code)).
					WithDriverMessage(pointer.SafeDeref(status.Message)).
					WithDriverRequestID(pointer.SafeDeref(response.Response.RequestId)).
					WithDriverResponse(response.Response)
			}
		}
	} else {
		return errors.ErrDriverSendFailed.
			WithDriverName(Name).
			WithDriverResponse(response)
	}

	return nil
}

package qcloud

import (
	"net/http"
	"time"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	smsv20210111 "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"go.gh.ink/smsutils/v3/errors"
	"go.gh.ink/smsutils/v3/model"
	"go.gh.ink/toolbox/expr"
)

type Client struct {
	Client      *smsv20210111.Client
	SmsSdkAPPID string
	// JSON
	Marshal   func(any) ([]byte, error)
	Unmarshal func([]byte, any) error
}
type Driver struct{}

func (d Driver) NewClient(params model.DriverClientParam) (model.Client, error) {
	// Check credential
	sid, skey := params.Credential[SecretID], params.Credential[SecretKey]
	if sid == "" || skey == "" {
		return Client{}, errors.ErrDriverCredentialInvalid
	}

	// Construct qcloud client config
	clientCredential := common.NewCredential(
		sid,
		skey,
	)

	// Create client profile
	cpf := profile.NewClientProfile()

	// Set qcloud endpoint
	cpf.HttpProfile.Endpoint = expr.Ternary(
		params.Credential[Endpoint] != "", params.Credential[Endpoint], DefaultEndpoint,
	)

	// Create QCloud client
	client, err := smsv20210111.NewClient(
		clientCredential,
		expr.Ternary(params.Credential[Region] != "", params.Credential[Region], DefaultRegion),
		cpf,
	)
	if err != nil {
		return nil, err
	}

	client.WithHttpTransport(responseTransport{next: http.DefaultTransport})
	return Client{
		Client:      client,
		SmsSdkAPPID: params.Credential[SmsSdkAppID],
		Marshal:     params.Marshal,
		Unmarshal:   params.Unmarshal,
	}, nil
}

var _ model.ClientCtx = Client{}
var _ model.DriverCtx = Driver{}

// NewClientCtx adds context support without changing the legacy driver contract.
func (d Driver) NewClientCtx(params model.DriverClientParam, options model.ClientOptions) (model.ClientCtx, error) {
	if options.Timeout < 0 {
		return nil, errors.ErrDriverCredentialInvalid.WithDriverName(Name)
	}
	if params.Credential[SmsSdkAppID] == "" {
		return nil, errors.ErrDriverCredentialInvalid.WithDriverName(Name)
	}

	legacy, err := d.NewClient(params)
	if err != nil {
		return nil, err
	}
	c := legacy.(Client)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = expr.Ternary(params.Credential[Endpoint] != "", params.Credential[Endpoint], DefaultEndpoint)
	if options.Timeout > 0 {
		cpf.HttpProfile.ReqTimeout = int((options.Timeout + time.Second - 1) / time.Second)
	}
	cpf.UnsafeRetryOnConnectionFailure = false
	c.Client.WithProfile(cpf)
	if options.Transport != nil {
		c.Client.WithHttpTransport(responseTransport{next: options.Transport})
	}
	return c, nil
}

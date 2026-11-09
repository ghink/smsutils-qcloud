package qcloud

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"go.gh.ink/smsutils/v3/errors"
)

// The SDK dereferences the decoded envelope while parsing errors. Validate it
// before handing untrusted response bytes to the SDK (JSON null must not panic).
type responseTransport struct{ next http.RoundTripper }

func (t responseTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := t.next.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if res == nil || res.Body == nil {
		return nil, errors.ErrDriverSendFailed.WithDriverName(Name)
	}
	raw, err := io.ReadAll(io.LimitReader(res.Body, 1024*1024+1))
	res.Body.Close()
	if err != nil {
		return nil, err
	}
	if len(raw) > 1024*1024 {
		return nil, errors.ErrDriverSendFailed.WithDriverName(Name)
	}
	var envelope struct{ Response json.RawMessage }
	if json.Unmarshal(raw, &envelope) != nil || len(envelope.Response) == 0 || bytes.Equal(bytes.TrimSpace(envelope.Response), []byte("null")) || bytes.TrimSpace(envelope.Response)[0] != '{' {
		return nil, errors.ErrDriverSendFailed.WithDriverName(Name)
	}
	res.Body = io.NopCloser(bytes.NewReader(raw))
	return res, nil
}

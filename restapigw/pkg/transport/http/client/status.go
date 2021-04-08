package client

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/core"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/errors"
)

// ===== [ Constants and Variables ] =====

const (
	// MWNamespace - HTTP Middleware 식별자
	MWNamespace = "mw-http"
)

var (
	// ErrInvalidStatusCode - Response의 StatusCode 가 200이나 201이 아닌 경우에 반환할 기본 오류
	ErrInvalidStatusCode = errors.New("Invalid status code")
)

// ===== [ Types ] =====

// HTTPStatusHandler - Response Code 처리를 위한 함수 정의
type HTTPStatusHandler func(context.Context, *http.Response) (*http.Response, error)

// HTTPResponseError - DetailedHTTPStatusHandler 처리 후에 반환되는 오류 구조 정의
// - responseError interface 구현체
type HTTPResponseError struct {
	Code int    `json:"http_status_code"`
	Msg  string `json:"http_body,omitempty"`
	Path string `json:"http_path"`
	name string
}

// ===== [ Implementations ] =====

// Error - 발생한 오류 메시지 반환
func (r HTTPResponseError) Error() string {
	return r.Msg
}

// Name - 발생한 오류 명 반환
func (r HTTPResponseError) Name() string {
	return r.name
}

// StatusCode - 발생한 오류 코드 반환
func (r HTTPResponseError) StatusCode() int {
	return r.Code
}

// ===== [ Private Functions ] =====

// ===== [ Public Functions ] =====

// DefaultHTTPStatusHandler - Request/Response 기준으로 StatusCode를 처리하는 기본 HTTPStatusHandler
func DefaultHTTPStatusHandler(ctx context.Context, resp *http.Response) (*http.Response, error) {
	// 오류 검증
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		msg, err := core.GetResponseString(resp)
		if err != nil {
			return nil, ErrInvalidStatusCode
		}

		// 오류가 발생했지만 Response가 존재하는 경우는 Response 반환
		if msg != "" {
			return resp, core.NewWrappedError(resp.StatusCode, msg, ErrInvalidStatusCode)
		}

		// 오류 반환
		return nil, ErrInvalidStatusCode
	}

	// 정상 반환
	return resp, nil
}

// NoOpHTTPStatusHandler - NO-OP 처리가 설정된 경우의 HTTPSTatusHandler (별도 처리가 필요없는 경우)
func NoOpHTTPStatusHandler(_ context.Context, resp *http.Response) (*http.Response, error) {
	return resp, nil
}

// DetailedHTTPStatusHandler - Request/Response 기준으로 발생한 오류에 정보를 상세하게 처리하기 위한 HTTPStatusHandler
func DetailedHTTPStatusHandler(next HTTPStatusHandler, name string) HTTPStatusHandler {
	return func(ctx context.Context, resp *http.Response) (*http.Response, error) {
		if r, err := next(ctx, resp); err == nil {
			return r, nil
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			body = []byte{}
		}
		resp.Body.Close()
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		// 오류가 존재하는 경우는 오류 정보 구성하고 Response도 같이 반환
		return resp, HTTPResponseError{
			Code: resp.StatusCode,
			Msg:  string(body),
			Path: resp.Request.URL.Path,
			name: name,
		}
	}
}

// GetHTTPStatusHandler - Backend를 호출한 후의 Response State 처리
// - "mw-http" Middleware 설정에 "return_error_details" 설정이 된 경우에는 DetailedHTTPStatusHandler를 사용 (내부적으로 DefaultHTTPStatusHandler 호출)
// - 그외는 DefaultHTTPStatusHandler 사용
func GetHTTPStatusHandler(bConf *config.BackendConfig) HTTPStatusHandler {
	if e, ok := bConf.Middleware[MWNamespace]; ok {
		if m, ok := e.(config.MWConfig); ok {
			if v, ok := m["return_error_details"]; ok {
				if b, ok := v.(string); ok && b != "" {
					return DetailedHTTPStatusHandler(DefaultHTTPStatusHandler, b)
				}
			}
		}
	}
	return DefaultHTTPStatusHandler
}

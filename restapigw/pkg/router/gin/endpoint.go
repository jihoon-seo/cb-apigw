package gin

import (
	"context"
	"fmt"
	"net/textproto"
	"strings"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/core"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/proxy"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/router"
	"github.com/gin-gonic/gin"
)

// ===== [ Constants and Variables ] =====

const (
	requestParamsAsterisk = "*"
)

// ===== [ Types ] =====

// responseError - Response에서 발생한 오류와 상태코드를 관리하는 오류
type responseError interface {
	error
	StatusCode() int
}

// HandlerFactory - 지정된 Endpoint 설정과 Proxy 를 기반으로 동작하는 Gin Framework handler 팩토리 정의
type HandlerFactory func(*config.EndpointConfig, proxy.Proxy) gin.HandlerFunc

// ===== [ Implementations ] =====

// ===== [ Private Functions ] =====

// ===== [ Public Functions ] =====

// EndpointHandler - 지정된 Endpoint 설정과 Proxy를 연계 호출하는 Gin Framework handler 생성
func EndpointHandler(eConf *config.EndpointConfig, proxy proxy.Proxy) gin.HandlerFunc {
	return CustomErrorEndpointHandler(eConf, proxy, router.DefaultToHTTPError)
}

// NewRequest - 지정한 Header 정보와 Query string 정보를 반영하는 Gin Context 기반의 Request 생성
func NewRequest(headersToSend []string) func(*gin.Context, []string) *proxy.Request {
	if len(headersToSend) == 0 {
		headersToSend = router.HeadersToSend
	}

	return func(c *gin.Context, queryString []string) *proxy.Request {
		params := make(map[string]string, len(c.Params))
		for _, param := range c.Params {
			params[strings.Title(param.Key)] = param.Value
		}

		headers := make(map[string][]string, 2+len(headersToSend))

		for _, k := range headersToSend {
			if k == requestParamsAsterisk {
				headers = c.Request.Header
				break
			}

			if h, ok := c.Request.Header[textproto.CanonicalMIMEHeaderKey(k)]; ok {
				headers[k] = h
			}
		}

		headers["X-Forwarded-For"] = []string{c.ClientIP()}
		// 전달되는 Header에 "User-Agent" 가 존재하지 않는 경우는 Router의 User-Agent 사용
		if _, ok := headers["User-Agent"]; !ok {
			headers["User-Agent"] = router.UserAgentHeaderValue
		} else {
			headers["X-Forwarded-Via"] = router.UserAgentHeaderValue
		}

		query := make(map[string][]string, len(queryString))
		queryValues := c.Request.URL.Query()
		for i := range queryString {
			if queryString[i] == requestParamsAsterisk {
				query = c.Request.URL.Query()
				break
			}

			if v, ok := queryValues[queryString[i]]; ok && len(v) > 0 {
				query[queryString[i]] = v
			}
		}

		return &proxy.Request{
			Method:  c.Request.Method,
			Query:   query,
			Body:    c.Request.Body,
			Params:  params,
			Headers: headers,
		}
	}
}

// CustomErrorEndpointHandler - 지정한 Endpoint 설정과 수행할 Proxy 정보 및 오류 처리를 위한 Gin Framework handler 생성
func CustomErrorEndpointHandler(eConf *config.EndpointConfig, proxy proxy.Proxy, errF router.ToHTTPError) gin.HandlerFunc {
	cacheControlHeaderValue := fmt.Sprintf("public, max-age=%d", int(eConf.CacheTTL.Seconds()))
	isCacheEnabled := eConf.CacheTTL.Seconds() != 0
	requestGenerator := NewRequest(eConf.HeadersToPass)
	render := getRender(eConf)

	return func(c *gin.Context) {
		requestCtx, cancel := context.WithTimeout(c, eConf.Timeout)
		c.Header(core.AppHeaderName, fmt.Sprintf("Version %s", core.AppVersion))
		response, err := proxy(requestCtx, requestGenerator(c, eConf.QueryString))
		select {
		case <-requestCtx.Done():
			if err == nil {
				err = router.ErrInternalError
			}
		default:
		}

		complete := router.HeaderIncompleteResponseValue
		if response != nil && len(response.Data) > 0 {
			if response.IsComplete {
				complete = router.HeaderCompleteResponseValue
				if isCacheEnabled {
					c.Header("Cache-Control", cacheControlHeaderValue)
				}
			}

			for k, vs := range response.Metadata.Headers {
				for _, v := range vs {
					c.Writer.Header().Add(k, v)
				}
			}
		}
		c.Header(router.CompleteResponseHeaderName, complete)

		if err != nil {
			// Proxy 처리 중에 발생한 오류들을 Header로 설정
			c.Header(router.MessageResponseHeaderName, err.Error())

			c.Error(err)

			// Response가 없는 경우의 상태 코드 설정
			if response == nil {
				if t, ok := err.(responseError); ok {
					c.Status(t.StatusCode())
				} else {
					c.Status(errF(err))
				}
				cancel()
				return
			}
		} else {
			c.Header(router.MessageResponseHeaderName, "OK")
		}

		render(c, response)
		cancel()
	}
}

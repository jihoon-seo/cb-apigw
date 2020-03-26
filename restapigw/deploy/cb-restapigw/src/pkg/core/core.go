// Package core - Defines variables/constants and provides utilty functions
package core

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

// ===== [ Constants and Variables ] =====

const (
	// AppName - 어플리케이션 명
	AppName = "cb-restapigw"
	// AppVersion - 어플리케이션 버전
	AppVersion = "0.1.0"
	// AppHeaderName - 어플리케이션 식별을 위한 Header 관리 명
	AppHeaderName = "X-CB-RESTAPIGW"
	// AppUserAgent - Backend 전달에 사용할 User Agent Header 값
	AppUserAgent = AppName + " version " + AppVersion
	// CollectionTag - Backend의 Array를 Json 객체의 데이터로 반환 처리를 위한 Tag Name
	CollectionTag = "collection"
	// WrappingTag = Backend의 Array 직접 반환 처리를 위한 Tag Name
	WrappingTag = "!!wrapping!!"
)

// ===== [ Types ] =====

// WrappedError - 원본 오류를 관리하는 오류 형식
type WrappedError struct {
	code          int
	message       string
	originalError error
}

// ===== [ Implementations ] =====

// Code - Wrapping된 오류 코드 반환
func (we WrappedError) Code() int {
	return we.code
}

// Error - 오류 메시지 반환
func (we WrappedError) Error() string {
	return fmt.Sprintf("%d, %s", we.code, we.message)
}

// GetError - 원본 오류 반환
func (we WrappedError) GetError() error {
	return we.originalError
}

// ===== [ Private Functions ] =====

// ===== [ Public Functions ] =====

// NewWrappedError - 원본 오류를 관리하는 오류 생성
func NewWrappedError(code int, message string, originalError error) error {
	return WrappedError{
		code:          code,
		message:       message,
		originalError: originalError,
	}
}

// GetStrings - 지정된 맵 데이터에서 지정된 이름에 해당하는 데이터를 []string 으로 반환
func GetStrings(data map[string]interface{}, name string) []string {
	result := []string{}
	if datas, ok := data[name]; ok {
		if data, ok := datas.([]interface{}); ok {
			for _, val := range data {
				if strVal, ok := val.(string); ok {
					result = append(result, strVal)
				}
			}
		}
	}
	return result
}

// GetString - 지정된 맵 데이터에서 지정한 키에 해당하는 데이터를 string 으로 반환
func GetString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// GetBool - 지정된 맵 데이터에서 지정한 키에 해당하는 데이터를 bool 으로 반환
func GetBool(data map[string]interface{}, key string) bool {
	if val, ok := data[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// GetInt64 - 지정된 맵 데이터에서 지정한 키에 해당하는 데이터를 int64 로 반환
func GetInt64(data map[string]interface{}, key string) int64 {
	if val, ok := data[key]; ok {
		switch i := val.(type) {
		case int64:
			return i
		case int:
			return int64(i)
		case float64:
			return int64(i)
		}
	}
	return 0
}

// ContainsString returns true if a string is present in a iteratee.
func ContainsString(s []string, v string) bool {
	for _, vv := range s {
		if vv == v {
			return true
		}
	}
	return false
}

// GetResponseString - http.Response Body를 문자열로 반환
func GetResponseString(resp *http.Response) (string, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	defer func() {
		resp.Body.Close()
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	}()

	return string(body), nil
}

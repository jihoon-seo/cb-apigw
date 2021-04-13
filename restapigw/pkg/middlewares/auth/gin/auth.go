// Package gin - API Gateway 접근인증을 위한 GIN 과 연동되는 기능 제공
package gin

import (
	"bytes"
	"net/http"
	"strings"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/errors"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/proxy"
	ginRouter "github.com/cloud-barista/cb-apigw/restapigw/pkg/router/gin"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

// ===== [ Constants and Variables ] =====

var (
	// MWNamespace - Middleware 설정 식별자
	MWNamespace = "mw-auth"
)

// ===== [ Types ] =====

// AuthMiddleware - HMAC Auth 처리가 적용된 Handler Func 반환 형식
type AuthMiddleware func(gin.HandlerFunc) gin.HandlerFunc

// Config - AUTH 운영을 위한 설정 구조
type Config struct {
	// HMAC 구성을 위한 보안 키
	SecureKey string `yaml:"secure_key"`
	// 접속을 허용하는 ID 리스트
	AcessIds []string `yaml:"access_ids"`
}

// ===== [ Implementations ] =====

// ===== [ Private Functions ] =====

// parseConfig - HTTPCache 운영을 위한 Configuration parsing 처리
func parseConfig(mwConf config.MWConfig) *Config {
	conf := new(Config)
	tmp, ok := mwConf[MWNamespace]
	if !ok {
		return nil
	}

	buf := new(bytes.Buffer)
	yaml.NewEncoder(buf).Encode(tmp)
	if err := yaml.NewDecoder(buf).Decode(conf); err != nil {
		return nil
	}

	return conf
}

// validateToken - Header로 전달된 Auth Token 검증
func validateToken(conf *Config, req *http.Request) error {
	token := req.Header.Get("Authorization")
	if token == "" {
		return errors.New("Authorization token not found")
	}

	token = strings.Replace(token, "Bearer ", "", 1)

	return ValidateToken(conf.SecureKey, token, conf.AcessIds)
}

// tokenValidator - Auth Token Validation 처리를 수행하는 Route Handler 구성
func tokenValidator(conf *Config, logger logging.Logger) AuthMiddleware {
	return func(next gin.HandlerFunc) gin.HandlerFunc {
		return func(c *gin.Context) {
			if err := validateToken(conf, c.Request); err != nil {
				logger.Error("[HMAC AUTH] Unauthorized accessing")
				c.AbortWithError(http.StatusUnauthorized, err)
				return
			}

			next(c)
		}
	}
}

// ===== [ Public Functions ] =====

// HandlerFactory - Auth 기능을 수행하는 Route Handler Factory 구성
func HandlerFactory(next ginRouter.HandlerFactory, logger logging.Logger) ginRouter.HandlerFactory {
	return func(eConf *config.EndpointConfig, p proxy.Proxy) gin.HandlerFunc {
		handlerFunc := next(eConf, p)

		conf := parseConfig(eConf.Middleware)
		if conf != nil {
			handlerFunc = tokenValidator(conf, logger)(handlerFunc)
		}

		return handlerFunc
	}
}

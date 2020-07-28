// Package jwt -
package jwt

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/jwt/provider"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/render"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/oauth2"
)

// ===== [ Constants and Variables ] =====

const (
	bearer = "bearer"
)

// ===== [ Types ] =====
type (
	// Handler - JWT 처리와 연계되는 구조
	Handler struct {
		Guard Guard
	}
)

// ===== [ Implementations ] =====

// Login - JWT Token 기반으로 인증 처리
func (h *Handler) Login(conf config.CredentialsConfig, logger logging.Logger) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		accessToken, err := extractAccessToken(req)

		if err != nil {
			logger.WithError(err).Debug("[ADMIN API] failed to extract access token")
		}

		httpClient := getClient(accessToken)
		factory := provider.Factory{}
		p := factory.Build(req.URL.Query().Get("provider"), conf)

		verified, err := p.Verify(req, httpClient)

		if nil != err || !verified {
			logger.WithError(err).Debug(err.Error())
			render.JSON(rw, http.StatusUnauthorized, err.Error())
			return
		}

		if 0 == h.Guard.Timeout {
			h.Guard.Timeout = time.Hour
		}

		claims, err := p.GetClaims(httpClient)
		if nil != err {
			render.JSON(rw, http.StatusBadRequest, err.Error())
			return
		}

		token, err := IssueAdminToken(h.Guard.SigningMethod, claims, h.Guard.Timeout)
		if nil != err {
			render.JSON(rw, http.StatusUnauthorized, "problem issuing JWT")
			return
		}

		render.JSON(rw, http.StatusOK, token)
	}
}

// Refresh - 현재 JWT에 대한 Refresh 처리
func (h *Handler) Refresh() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		parser := Parser{h.Guard.ParserConfig}
		token, _ := parser.ParseFromRequest(req)
		claims := token.Claims.(jwt.MapClaims)

		origIat := int64(claims["iat"].(float64))

		if origIat < time.Now().Add(-h.Guard.MaxRefresh).Unix() {
			render.JSON(rw, http.StatusUnauthorized, "token is expired")
			return
		}

		// Refresh 용 토큰 생성
		newToken := jwt.New(jwt.GetSigningMethod(h.Guard.SigningMethod.Alg))
		newClaims := newToken.Claims.(jwt.MapClaims)

		for key := range claims {
			newClaims[key] = claims[key]
		}

		expire := time.Now().Add(h.Guard.Timeout)
		newClaims["sub"] = claims["sub"]
		newClaims["exp"] = expire.Unix()
		newClaims["iat"] = origIat

		// 현재는 HSxxx 형식 알고리즘만 지원
		tokenString, err := newToken.SignedString([]byte(h.Guard.SigningMethod.Key))
		if err != nil {
			render.JSON(rw, http.StatusUnauthorized, "create JWT Token failed")
			return
		}

		render.JSON(rw, http.StatusOK, render.M{
			"token":  tokenString,
			"type":   "Bearer",
			"expire": expire.Format(time.RFC3339),
		})
	}
}

// ===== [ Private Functions ] =====

// extractAccessToken - Request 정보에서 인증 정보 추출
func extractAccessToken(req *http.Request) (string, error) {
	// Access Key들을 검증하는 OAuth 활용
	authHeaderValue := req.Header.Get("Authorization")
	parts := strings.Split(authHeaderValue, " ")
	if len(parts) < 2 {
		return "", errors.New("attempted access with malformed header, no auth header found")
	}

	if strings.ToLower(parts[0]) != bearer {
		return "", errors.New("bearer token malformed")
	}

	return parts[1], nil
}

// getClient - 지정한 토큰을 처리하기 위한 HTTP Client 구성
func getClient(token string) *http.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	return oauth2.NewClient(ctx, ts)
}

// ===== [ Public Functions ] =====

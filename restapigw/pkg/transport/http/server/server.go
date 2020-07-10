// Package server - API G/W 및 Admin 운영을 위한 Server 기능 제공 패키지
package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
)

// ===== [ Constants and Variables ] =====

var (
	// ErrPrivateKey - TLS 설정에서 사용할 Private Key 미 설정 오류
	ErrPrivateKey = errors.New("private key not defined")
	// ErrPublicKey - TLS 설정에서 사용할 Public Key 미 설정 오류
	ErrPublicKey = errors.New("public key not defined")

	defaultCurves = []tls.CurveID{
		tls.CurveP521,
		tls.CurveP384,
		tls.CurveP256,
	}
	defaultCipherSuites = []uint16{
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
	}
	versions = map[string]uint16{
		"SSL3.0": tls.VersionSSL30,
		"TLS10":  tls.VersionTLS10,
		"TLS11":  tls.VersionTLS11,
		"TLS12":  tls.VersionTLS12,
	}
)

// ===== [ Types ] =====
// ===== [ Implementations ] =====
// ===== [ Private Functions ] =====

func parseTLSVersion(key string) uint16 {
	if v, ok := versions[key]; ok {
		return v
	}
	return tls.VersionTLS12
}

func parseCurveIDs(conf *config.TLSConfig) []tls.CurveID {
	l := len(conf.CurvePreferences)
	if l == 0 {
		return defaultCurves
	}

	curves := make([]tls.CurveID, len(conf.CurvePreferences))
	for i := range curves {
		curves[i] = tls.CurveID(conf.CurvePreferences[i])
	}
	return curves
}

func parseCipherSuites(conf *config.TLSConfig) []uint16 {
	l := len(conf.CipherSuites)
	if l == 0 {
		return defaultCipherSuites
	}

	cs := make([]uint16, l)
	for i := range cs {
		cs[i] = uint16(conf.CipherSuites[i])
	}
	return cs
}

// parseTLSConfig - 서비스 설정에 지정된 TLS 설정을 기준으로 tls 모듈에 대한 설정 반환
func parseTLSConfig(tlsConf *config.TLSConfig) *tls.Config {
	if tlsConf == nil {
		return nil
	}
	if tlsConf.IsDisabled {
		return nil
	}

	return &tls.Config{
		MinVersion:               parseTLSVersion(tlsConf.MinVersion),
		MaxVersion:               parseTLSVersion(tlsConf.MaxVersion),
		CurvePreferences:         parseCurveIDs(tlsConf),
		PreferServerCipherSuites: tlsConf.PreferServerCipherSuites,
		CipherSuites:             parseCipherSuites(tlsConf),
	}
}

// ===== [ Public Functions ] =====

// NewServer - 지정한 설정과 http handler 기준으로 동작하는 http server 반환
func NewServer(sConf config.ServiceConfig, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf(":%d", sConf.Port),
		Handler:           handler,
		ReadTimeout:       sConf.ReadTimeout,
		WriteTimeout:      sConf.WriteTimeout,
		ReadHeaderTimeout: sConf.ReadHeaderTimeout,
		IdleTimeout:       sConf.IdleTimeout,
		TLSConfig:         parseTLSConfig(sConf.TLS),
	}
}

// RunServer - 지정된 Context와 설정 및 Handler 기반으로 동작하는 HTTP Server 구동
func RunServer(ctx context.Context, sConf config.ServiceConfig, handler http.Handler, log logging.Logger) error {
	done := make(chan error)

	s := NewServer(sConf, handler)

	if s.TLSConfig == nil {
		go func() {
			done <- s.ListenAndServe()
		}()
	} else {
		if sConf.TLS.PublicKey == "" {
			return ErrPublicKey
		}
		if sConf.TLS.PrivateKey == "" {
			return ErrPrivateKey
		}
		go func() {
			done <- s.ListenAndServeTLS(sConf.TLS.PublicKey, sConf.TLS.PrivateKey)
		}()
	}

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return s.Shutdown(context.Background())
	}
}

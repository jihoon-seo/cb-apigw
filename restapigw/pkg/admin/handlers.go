// Package admin -
package admin

import (
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/core"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/render"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====
// ===== [ Implementations ] =====
// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// Home handler is just a nice home page message
func Home() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, http.StatusOK, "Welcome to "+core.AppName+" - ADMIN Manager")
	}
}

// RedirectHTTPS - HTTP로 전달된 Request를 HTTPS로 전달
func RedirectHTTPS(port int) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		host, _, _ := net.SplitHostPort(req.Host)

		target := url.URL{
			Scheme: "https",
			Host:   fmt.Sprintf("%s:%v", host, port),
			Path:   req.URL.Path,
		}
		if len(req.URL.RawQuery) > 0 {
			target.RawQuery += "?" + req.URL.RawQuery
		}
		logging.GetLogger().Infof("redirect to: %s", target.String())
		http.Redirect(w, req, target.String(), http.StatusTemporaryRedirect)
	})
}

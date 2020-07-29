// Package admin -
package admin

import (
	"fmt"
	"net/http"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/api"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/core"
	ginAdapter "github.com/cloud-barista/cb-apigw/restapigw/pkg/core/adapters/gin"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/errors"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/render"
	"go.opencensus.io/trace"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====

type (
	// APIHandler - Admin API 운영을 위한 REST Handler 정보 형식
	APIHandler struct {
		configurationChan chan<- api.ConfigurationMessage

		Configs *api.Configuration
	}
)

// ===== [ Implementations ] =====

// exists - 지정한 Endpoint Defintion이 존재하는지 검증
func (ah *APIHandler) exists(def *config.EndpointConfig) (bool, error) {
	for _, storedDef := range ah.Configs.Definitions {
		if storedDef.Name == def.Name {
			return true, api.ErrAPINameExists
		}

		if storedDef.Endpoint == def.Endpoint {
			return true, api.ErrAPIListenPathExists
		}
	}

	return false, nil
}

// findByName - 지정한 이름의 Endpoint Definition이 존재하는지 검증
func (ah *APIHandler) findByName(name string) *config.EndpointConfig {
	for _, storedDef := range ah.Configs.Definitions {
		if storedDef.Name == name {
			return storedDef
		}
	}

	return nil
}

// findByListenPath - 지정한 Path의 Endpoint Definition이 존재하는 검증
func (ah *APIHandler) findByListenPath(listenPath string) *config.EndpointConfig {
	for _, cfg := range ah.Configs.Definitions {
		if cfg.Endpoint == listenPath {
			return cfg
		}
	}

	return nil
}

// Get - 전체 Definition 정보 반환
func (ah *APIHandler) Get() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		_, span := trace.StartSpan(req.Context(), "definitions.GetAll")
		defer span.End()

		if nil == ah.Configs.Definitions {
			// API Definition이 없을 경우는 빈 JSON Array 처리 (ID 기준)
			render.JSON(rw, http.StatusOK, []int{})
			return
		}

		//render.JSON(rw, http.StatusOK, core.ConvertStructForJSON(ah.Configs.Definitions))
		render.JSON(rw, http.StatusOK, ah.Configs.Definitions)
	}
}

// GetBy - Request의 name 정보 기준으로 Definition 정보 반환
func (ah *APIHandler) GetBy() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		name := ginAdapter.URLParam(req, "name")
		_, span := trace.StartSpan(req.Context(), "definition.FindByName")
		def := ah.findByName(name)
		span.End()

		if nil == def {
			errors.Handler(rw, req, api.ErrAPIDefinitionNotFound)
			return
		}

		render.JSON(rw, http.StatusOK, def)
	}
}

// PutBy - Request의 name 정보를 기준으로 Definition 정보 갱신
func (ah *APIHandler) PutBy() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		var err error

		name := ginAdapter.URLParam(req, "name")
		_, span := trace.StartSpan(req.Context(), "definition.FindByName")
		def := ah.findByName(name)
		span.End()

		if nil == def {
			errors.Handler(rw, req, api.ErrAPIDefinitionNotFound)
			return
		}

		err = core.JSONDecode(req.Body, def)
		if nil != err {
			errors.Handler(rw, req, err)
			return
		}

		isValid, err := def.Validate()
		if !isValid && nil != err {
			errors.Handler(rw, req, errors.NewWithCode(http.StatusBadRequest, err.Error()))
			return
		}

		// TODO: Plugin Validation

		// 동일한 경로가 다른 Definition 이름으로 등록되어있는 경우 검증
		_, span = trace.StartSpan(req.Context(), "repo.FindByListenPath")
		existingDef := ah.findByListenPath(def.Endpoint)
		span.End()

		if nil != existingDef && existingDef.Name != def.Name {
			errors.Handler(rw, req, api.ErrAPIListenPathExists)
			return
		}

		// Definition 갱신 채널 처리
		_, span = trace.StartSpan(req.Context(), "repo.Update")
		ah.configurationChan <- api.ConfigurationMessage{
			Operation:     api.UpdatedOperation,
			Configuration: def,
		}
		span.End()

		rw.WriteHeader(http.StatusOK)
	}
}

// Post - Request의 body 정보를 기준으로 Definition 정보 추가
func (ah *APIHandler) Post() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		def := api.NewDefinition()

		err := core.JSONDecode(req.Body, def)
		if nil != err {
			errors.Handler(rw, req, err)
			return
		}

		isValid, err := def.Validate()
		if false == isValid && err != nil {
			errors.Handler(rw, req, errors.NewWithCode(http.StatusBadRequest, err.Error()))
			return
		}

		// TODO: Plugin Validation

		// 기존 정보가 존재하는지 검증 (Name 및 Endpoint Path, ...)
		_, span := trace.StartSpan(req.Context(), "definition.Exists")
		exists, err := ah.exists(def)
		span.End()

		if err != nil || exists {
			errors.Handler(rw, req, err)
			return
		}

		// Definition 갱신 채널 처리
		_, span = trace.StartSpan(req.Context(), "repo.Add")
		ah.configurationChan <- api.ConfigurationMessage{
			Operation:     api.AddedOperation,
			Configuration: def,
		}
		span.End()

		rw.Header().Add("Location", fmt.Sprintf("/apis/%s", def.Name))
		rw.WriteHeader(http.StatusCreated)
	}
}

// DeleteBy - Request의 name 정보를 기준으로 Definition 정보 삭제
func (ah *APIHandler) DeleteBy() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		_, span := trace.StartSpan(req.Context(), "repo.Remove")
		defer span.End()

		name := ginAdapter.URLParam(req, "name")
		cfg := ah.findByName(name)
		if cfg == nil {
			errors.Handler(rw, req, api.ErrAPIDefinitionNotFound)
			return
		}

		// Definition 갱신 채널 처리
		ah.configurationChan <- api.ConfigurationMessage{
			Operation:     api.RemovedOperation,
			Configuration: cfg,
		}

		rw.WriteHeader(http.StatusNoContent)
	}
}

// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// NewAPIHandler - 지정한 Configuration 변경을 설정한 Admin API Handler 인스턴스 생성
func NewAPIHandler(configurationChan chan<- api.ConfigurationMessage) *APIHandler {
	return &APIHandler{
		configurationChan: configurationChan,
	}
}

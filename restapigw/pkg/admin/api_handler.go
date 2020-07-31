// Package admin -
package admin

import (
	"fmt"
	"net/http"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/api"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/core"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/errors"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/render"
	"go.opencensus.io/trace"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====

type (
	// APIHandler - Admin API 운영을 위한 REST Handler 정보 형식
	APIHandler struct {
		configurationChan chan<- api.ChangeMessage

		Configs *api.Configuration
	}
)

// ===== [ Implementations ] =====

// GetDefinitions - 전체 Definition 정보 반환
func (ah *APIHandler) GetDefinitions() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		_, span := trace.StartSpan(req.Context(), "definitions.GetAll")
		defer span.End()

		if nil == ah.Configs.DefinitionMaps {
			// API Definition이 없을 경우는 빈 JSON Array 처리 (ID 기준)
			render.JSON(rw, http.StatusOK, []int{})
			return
		}

		render.JSON(rw, http.StatusOK, ah.Configs.DefinitionMaps)
	}
}

// UpdateDefinition - Request의 name 정보를 기준으로 Definition 정보 갱신
func (ah *APIHandler) UpdateDefinition() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		cm := &api.ConfigModel{}

		err := core.JSONDecode(req.Body, cm)
		if nil != err {
			errors.Handler(rw, req, err)
			return
		}

		_, span := trace.StartSpan(req.Context(), "definition.FindByName")
		def := ah.Configs.FindByName(cm.Source, cm.Definition.Name)
		span.End()

		if nil == def {
			errors.Handler(rw, req, api.ErrAPIDefinitionNotFound)
			return
		}

		isValid, err := cm.Definition.Validate()
		if !isValid && nil != err {
			errors.Handler(rw, req, errors.NewWithCode(http.StatusBadRequest, err.Error()))
			return
		}

		// TODO: Plugin Validation

		// 동일한 경로가 다른 Definition 이름으로 등록되어있는 경우 검증
		_, span = trace.StartSpan(req.Context(), "repo.FindByListenPath")
		existingDef := ah.Configs.FindByListenPath(cm.Source, cm.Definition.Endpoint)
		span.End()

		if nil != existingDef && existingDef.Name != cm.Definition.Name {
			errors.Handler(rw, req, api.ErrAPIListenPathExists)
			return
		}

		// Definition 갱신 채널 처리
		_, span = trace.StartSpan(req.Context(), "repo.Update")
		ah.configurationChan <- api.ChangeMessage{
			Source:     cm.Source,
			Operation:  api.UpdatedOperation,
			Definition: cm.Definition,
		}
		span.End()

		rw.WriteHeader(http.StatusOK)
	}
}

// AddDefinition - Request의 body 정보를 기준으로 Definition 정보 추가
func (ah *APIHandler) AddDefinition() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		cm := &api.ConfigModel{}

		err := core.JSONDecode(req.Body, cm)
		if nil != err {
			errors.Handler(rw, req, err)
			return
		}

		isValid, err := cm.Definition.Validate()
		if false == isValid && nil != err {
			errors.Handler(rw, req, errors.NewWithCode(http.StatusBadRequest, err.Error()))
			return
		}

		// TODO: Plugin Validation

		// 기존 정보가 존재하는지 검증 (Name 및 Endpoint Path, ...)
		_, span := trace.StartSpan(req.Context(), "definition.Exists")
		exists, err := ah.Configs.Exists(cm.Source, cm.Definition)
		span.End()

		if nil != err || exists {
			errors.Handler(rw, req, err)
			return
		}

		// Definition 갱신 채널 처리
		_, span = trace.StartSpan(req.Context(), "repo.Add")
		ah.configurationChan <- api.ChangeMessage{
			Source:     cm.Source,
			Operation:  api.AddedOperation,
			Definition: cm.Definition,
		}
		span.End()

		rw.Header().Add("Location", fmt.Sprintf("/apis/%s", cm.Definition.Name))
		rw.WriteHeader(http.StatusCreated)
	}
}

// RemoveDefinition - Request의 name 정보를 기준으로 Definition 정보 삭제
func (ah *APIHandler) RemoveDefinition() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		cm := &api.ConfigModel{}

		err := core.JSONDecode(req.Body, cm)
		if nil != err {
			errors.Handler(rw, req, err)
			return
		}

		_, span := trace.StartSpan(req.Context(), "repo.Remove")
		defer span.End()

		cm.Definition = ah.Configs.FindByName(cm.Source, cm.Definition.Name)
		if nil == cm.Definition {
			errors.Handler(rw, req, api.ErrAPIDefinitionNotFound)
			return
		}

		// Definition 갱신 채널 처리
		ah.configurationChan <- api.ChangeMessage{
			Source:     cm.Source,
			Operation:  api.RemovedOperation,
			Definition: cm.Definition,
		}

		rw.WriteHeader(http.StatusNoContent)
	}
}

// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// NewAPIHandler - 지정한 Configuration 변경을 설정한 Admin API Handler 인스턴스 생성
func NewAPIHandler(configurationChan chan<- api.ChangeMessage) *APIHandler {
	return &APIHandler{
		configurationChan: configurationChan,
	}
}

// Package api -
package api

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/errors"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
	"gopkg.in/yaml.v3"
)

// ===== [ Constants and Variables ] =====

const (
	file    = "file"
	cbStore = "cbstore"
)

var ()

// ===== [ Types ] =====
type (
	// GroupDefinitions - 리파지토리 Group에 저장된 API Definition 구조 (로드/저장용)
	GroupDefinitions struct {
		Definitions []*config.EndpointConfig `mapstructure:"definitions" yaml:"definitions"`
	}

	// DefinitionMap - 리파지토리의 API Definition 관리 정보 구조 (관리 및 클라이언트 연계용)
	DefinitionMap struct {
		Name        string                   `json:"name"`
		State       ConfigurationState       `json:"-"`
		Definitions []*config.EndpointConfig `json:"definitions"`
	}

	// Repository - Routing 정보 관리 기능을 제공하는 인터페이스 형식
	Repository interface {
		io.Closer

		FindAll() ([]*DefinitionMap, error)
		Write([]*DefinitionMap) error
	}

	// Watcher - Repository에서 API Definition 변경 여부 감시용
	Watcher interface {
		Watch(ctx context.Context, configurationChan chan<- RepoChangedMessage)
	}

	// Listener - 관리 중인 API Definition 변경 여부 감시용
	Listener interface {
		Listen(ctx context.Context, configurationChan <-chan ConfigChangedMessage)
	}
)

// ===== [ Implementations ] =====

// CheckDuplicates - 관리중인 Definitions에 대해 중복 검증
func (dm *DefinitionMap) CheckDuplicates(group string, eConf *config.EndpointConfig) bool {
	log := logging.GetLogger()

	for _, def := range dm.Definitions {
		if strings.EqualFold(dm.Name, group) && strings.EqualFold(def.Name, eConf.Name) {
			log.Warnf("Same Endpoint's Name [%s] exist on group [%s]. Endpoint's Name must be unique in group", eConf.Name, group)
			return true
		}
		if strings.EqualFold(def.Endpoint, eConf.Endpoint) {
			log.Warnf("Same Endpoint [%s] exist. Endpoint must be unique in all groups", eConf.Endpoint)
			return true
		}
	}
	return false
}

// ===== [ Private Functions ] =====

// parseEndpoint - 지정된 정보를 Definition 정보로 전환
func parseEndpoint(sConf *config.ServiceConfig, apiDef []byte) (*GroupDefinitions, error) {
	var apiConfigs *GroupDefinitions

	// API 정의들 Unmarshalling
	if err := yaml.Unmarshal(apiDef, &apiConfigs); nil != err {
		return nil, err
	}

	// 로드된 Endpoint 정보 재 구성
	for _, def := range apiConfigs.Definitions {
		// 서비스 설정에서 상속할 정보들 설정
		def.InheriteFromService(sConf)

		// 서비스 설정과 연계해서 처리할 정보들 설정
		if err := def.AdjustValues(sConf); nil != err {
			return nil, errors.Wrapf(err, "couldn't initialize api definition (adjust values): '%s'", def.Name)
		}
	}

	return apiConfigs, nil
}

// groupDefinitions - 출력을 위한 Group Definition 구조 반환
func groupDefinitions(dm *DefinitionMap) ([]byte, error) {
	sd := &GroupDefinitions{Definitions: dm.Definitions}
	data, err := yaml.Marshal(sd)
	if nil != err {
		return nil, err
	}
	return data, nil
}

// ===== [ Public Functions ] =====

// BuildRepository - 시스템 설정에 정의된 DSN(Data Group Name) 기준으로 저장소 구성
func BuildRepository(sConf *config.ServiceConfig, refreshTime time.Duration) (Repository, error) {
	log := logging.GetLogger()
	dsnURL, err := url.Parse(sConf.Repository.DSN)
	if nil != err {
		return nil, errors.Wrap(err, "Error parsing the DSN")
	}
	if dsnURL.Path == "" {
		return nil, errors.New("Path not found from DSN")
	}

	// File 모드인 경우는 상대경로 처리 검증
	if strings.EqualFold(dsnURL.Scheme, "file") && strings.EqualFold(dsnURL.Host, ".") {
		path, err := filepath.Abs(dsnURL.Host + dsnURL.Path)
		if nil != err {
			return nil, err
		}
		dsnURL.Path = path
	}

	switch dsnURL.Scheme {
	// CB-STORE (NutsDB or ETCD) 사용
	case cbStore:
		log.Debug("[REPOSITORY] CB-Store (NutsDB or ETCD) based configuration chosen")
		storeKey := dsnURL.Path

		log.WithField("key", storeKey).Debug("[REPOSITORY] Trying to load API configuration files")
		repo, err := NewCbStoreRepository(sConf, storeKey, refreshTime)
		if nil != err {
			return nil, errors.Wrap(err, "could not create a CB-Store repository")
		}
		return repo, nil
	// File(Memoery) 사용
	case file:
		log.Debug("[REPOSITORY] File system based configuration chosen")
		apiPath := fmt.Sprintf("%s/apis", dsnURL.Path)

		log.WithField("path", apiPath).Debug("[REPOSITORY] Trying to load API configuration files")
		repo, err := NewFileSystemRepository(sConf, apiPath)
		if nil != err {
			return nil, errors.Wrap(err, "could not create a file system repository")
		}
		return repo, nil
	default:
		return nil, errors.New("The selected scheme is not supported to load API definitions")
	}
}

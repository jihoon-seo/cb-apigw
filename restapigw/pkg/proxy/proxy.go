// Package proxy - REST API G/W 의 Reverse Proxy 기능 제공 패키지
package proxy

import (
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/core"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/proxy/balancer"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/router"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====
type (
	// Definition - Route 처리를 위한 Proxy Rule 구조 (이전 Endpoint)
	Definition struct {
		// TODO: Bypass
		// TODO: Output Encoding
		// TODO: Except QueryString
		// TODO: Except Header
		// TODO: Middleware to Plugin
		PreserveHost       bool               `mapstructure:"preserve_host" bson:"preserve_host" json:"preserve_host"`
		ListenPath         string             `mapstructure:"listen_path" bson:"listen_path" json:"listen_path" valid:"required~proxy.listen_path is required,urlpath"`
		Upstreams          *Upstreams         `mapstructure:"upstreams" bson:"upstreams" json:"upstreams"`
		InsecureSkipVerify bool               `mapstructure:"insecure_skip_verify" bson:"insecure_skip_verify" json:"insecure_skip_verify"`
		StripPath          bool               `mapstructure:"strip_path" bson:"strip_path" json:"strip_path"`
		AppendPath         bool               `mapstructure:"append_path" bson:"append_path" json:"append_path"`
		Methods            []string           `mapstructure:"methods" bson:"methods" json:"methods"`
		Hosts              []string           `mapstructure:"hosts" bson:"hosts" json:"hosts"`
		ForwardingTimeouts ForwardingTimeouts `mapstructure:"forwarding_timeouts" bson:"forwarding_timeouts" json:"forwarding_timeouts"`
	}

	// RouterDefinition - 내부적으로 Router 처리에 사용할 Proxy Rule과 Middleware 관리 구조
	RouterDefinition struct {
		*Definition
		middleware []router.Constructor
	}

	// Upstreams - Request가 전달될 대상 관리 정보 (이전 Backend의 컬랙션 - LoadBalancing 적용)
	Upstreams struct {
		Balancing string  `mapstructure:"balancing" bson:"balancing" json:"balancing"`
		Targets   Targets `mapstructure:"targets" bson:"targets" json:"targets"`
	}

	// Targets - Target 구조 컬랙션
	Targets []*Target

	// Target - Backend Service 관리 정보 (이전 Backend)
	Target struct {
		// TODO: Method?
		// TODO: Manipulation of result
		Target string `mapstructure:"target" bson:"target" json:"target"`
		Weight int    `mapstructure:"weight" bson:"weight" json:"weight"`
	}

	// ForwardingTimeouts - Rquest를 Backend Service로 전달할 떄 적용할 Timeout 관리 구조
	ForwardingTimeouts struct {
		DialTimeout           core.Duration `mapstructure:"dial_timeout" bson:"dial_timeout" json:"dial_timeout"`
		ResponseHeaderTimeout core.Duration `mapstructure:"response_header_timeout" bson:"response_header_timeout" json:"response_header_timeout"`
	}

	// // EndpointConfig - 서비스 라우팅에 사용할 설정 구조
// type EndpointConfig struct {
// 	// Bypass 처리 여부
// 	IsBypass bool

// 	// 클라이언트에 노출될 URL 패턴
// 	Endpoint string `mapstructure:"endpoint"`
// 	// Endpoint에 대한 HTTP 메서드 (GET, POST, PUT, etc) (기본값: GET)
// 	Method string `mapstructure:"method"`
// 	// Endpoint 처리 시간 (기본값: 서비스 값 사용)
// 	Timeout time.Duration `mapstructure:"timeout"`
// 	// GET 처리에 대한 캐시 TTL 기간 (기본값: 서비스 값 사용)
// 	CacheTTL time.Duration `mapstructure:"cache_ttl"`
// 	// 반환결과 처리에 사용할 인코딩
// 	OutputEncoding string `mapstructure:"output_encoding"`
// 	// Backend 에 전달되는 Query String에서 제외할 파라미터 Key 리스트
// 	ExceptQueryStrings []string `mapstructure:"except_querystrings"`
// 	// Backend 에 전달되는 Header에서 제외할 파라미터 Key 리스트
// 	ExceptHeaders []string `mapstructure:"except_headers"`
// 	// Endpoint 단위에서 적용할 Middleware 설정
// 	Middleware MWConfig `mapstructure:"middleware"`
// 	// Endpoint에서 호출할 Backend API 서버 호출/응답 처리 설정 리스트
// 	Backend []*BackendConfig `mapstructure:"backend"`
// }

// // BackendConfig - Backend API Server 연결과 응답 처리를 위한 설정 구조
// type BackendConfig struct {
// 	// Backend API Server의 Host URI (서비스에서 전역으로 설정한 경우는 생략 가능)
// 	Host []string `mapstructure:"host"`
// 	// Backend 처리 시간 (기본값: )
// 	Timeout time.Duration
// 	// Backend 호출에 사용할 HTTP Method
// 	Method string `mapstructure:"method"`
// 	// Backend 호출에 사용할 URL Patthern
// 	URLPattern string `mapstructure:"url_pattern"`
// 	// 인코딩 포맷
// 	Encoding string `mapstructure:"encoding"`
// 	// Backend 결과를 묶을 Group 명 (기본값: )
// 	Group string `mapstructure:"group"`
// 	// Backend 결과에서 생략할 필드명 리스트 (flatmap 적용 "." operation)
// 	Blacklist []string `mapstructure:"blacklist"`
// 	// Backend 결과에서 추출할 필드명 리스트 (flatmap 적용 "." operation)
// 	Whitelist []string `mapstructure:"whitelist"`
// 	// Backend 결과에서 필드명을 변경할 리스트 맵
// 	Mapping map[string]string `mapstructure:"mapping"`
// 	// Backend 결과가 컬랙션인지 여부
// 	IsCollection bool `mapstructure:"is_collection"`
// 	// Backend 결과가 컬랙션인 경우에 core.CollectionTag ("collection") 으로 JSON 포맷을 할 것인지 여부 (True 면 core.CollectionTag ("collection") 으로 JSON 전환, false면 Array 상태로 반환)
// 	WrapCollectionToJSON bool `mapstructure:"wrap_collection_to_json"`
// 	// Backend 결과 중에서 특정한 필드만 처리할 경우의 필드명
// 	Target string `mapstructure:"target"`
// 	// Backend 에서 동작할 Middleware 설정
// 	Middleware MWConfig `mapstructure:"middleware"`
// 	// HostSanitizationDisabled - host 정보의 정제작업 비활성화 여부
// 	HostSanitizationDisabled bool `mapstructure:"disable_host_sanitize"`
// 	// API 호출의 응답을 파싱하기 위한 디코더 (내부 사용)
// 	Decoder encoding.Decoder `json:"-"`
// 	// URLPattern에서 파라미터 변환에 사용할 키 관리 (내부 사용)
// 	URLKeys []string
// }
)

// ===== [ Implementations ] =====

// Middleware - RouteDefinition에 지정된 미들웨어들을 반환
func (rd *RouterDefinition) Middleware() []router.Constructor {
	return rd.middleware
}

// AddMiddleware - 지정한 미들웨어 정보를 관리 중인 미들웨어 스택에 추가
func (rd *RouterDefinition) AddMiddleware(m router.Constructor) {
	rd.middleware = append(rd.middleware, m)
}

// Validate - Proxy Definition 정보 검증
func (d *Definition) Validate() (bool, error) {
	return govalidator.ValidateStruct(d)
}

// IsBalancerDefined - Proxy Definition에서 Target 정의에 따른 Balaner 여부 반환
func (d *Definition) IsBalancerDefined() bool {
	return d.Upstreams != nil && d.Upstreams.Targets != nil && len(d.Upstreams.Targets) > 0
}

// ToBalancerTargets - 예상되는 형식의 Balancer에 해당하는 Target들 반환
func (t Targets) ToBalancerTargets() []*balancer.Target {
	var balancerTargets []*balancer.Target
	for _, t := range t {
		balancerTargets = append(balancerTargets, &balancer.Target{
			Target: t.Target,
			Weight: t.Weight,
		})
	}
	return balancerTargets
}

// ===== [ Private Functions ] =====

// init - 패키지 초기화
func init() {
	// Initialize Custom Validators
	govalidator.CustomTypeTagMap.Set("urlpath", func(i interface{}, o interface{}) bool {
		s, ok := i.(string)
		if !ok {
			return false
		}

		return strings.Index(s, "/") == 0
	})
}

// ===== [ Public Functions ] =====

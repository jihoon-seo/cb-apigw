package router

import "regexp"

// ===== [ Constants and Variables ] =====

const (
	parameterMatchRule = `\{([^/}]+)\}`
)

// ===== [ Types ] =====

type (
	// ListenPathParameterNameExtractor - Listen Path에서 파리미터 이름을 추출하기 위한 정규식 관리 구조
	ListenPathParameterNameExtractor struct {
		reg *regexp.Regexp
	}
)

// ===== [ Implementations ] =====

// Extract - 지정한 경로에서 파라미터 이름을 추출해서 반환
func (l *ListenPathParameterNameExtractor) Extract(lp string) []string {
	submatches := l.reg.FindAllStringSubmatch(lp, -1)
	result := make([]string, 0, len(submatches))

	for _, submatch := range submatches {
		result = append(result, submatch[1])
	}

	return result
}

// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// NewListenPathParamNameExtractor - Listening 하는 Path 정보에서 파라미터로 사용되는 이름을 추출하는 인스턴스 생성
func NewListenPathParamNameExtractor() *ListenPathParameterNameExtractor {
	return &ListenPathParameterNameExtractor{regexp.MustCompile(parameterMatchRule)}
}

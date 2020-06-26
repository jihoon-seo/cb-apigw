package router

import "regexp"

// ===== [ Constants and Variables ] =====
const (
	matchRule = `(\/\*(.+)?)`
)

// ===== [ Types ] =====

type (
	ListenPathMatcher struct {
		reg *regexp.Regexp
	}
)

// ===== [ Implementations ] =====

// Match - 지정한 경로가 관리하고 있는 매칭 룰에 해당하는지 반환
func (m *ListenPathMatcher) Match(listenPath string) bool {
	return m.reg.MatchString(listenPath)
}

// Extract - 지정한 경로에서 관리하고 있는 매칭 룰에 해당하는 부분을 추출해서 반환
func (m *ListenPathMatcher) Extract(listenPath string) string {
	return m.reg.ReplaceAllString(listenPath, "")
}

// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// NewListenPathMatcher - 수신된 경로를 특정하기 위한 정규식을 처리하는 Matcher 인스턴스 생성
func NewListenPathMatcher() *ListenPathMatcher {
	return &ListenPathMatcher{regexp.MustCompile(matchRule)}
}

// Package observability - 관찰 기능을 제공하는 패키지
package observability

import "go.opencensus.io/tag"

// ===== [ Constants and Variables ] =====

var (
	// KeyListenPath - Opencensus 연계를 위한 ListenPath 관련 태그 키
	KeyListenPath, _ = tag.NewKey("path")
	// KeyUpstreamPath - Opencensus 연계를 위한 UpstreamPath 관련 태크 키
	KeyUpstreamPath, _ = tag.NewKey("upstream_path")
)

// ===== [ Types ] =====
// ===== [ Implementations ] =====
// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// Package core - Defines variables/constants and provides utilty functions
package core

// ===== [ Constants and Variables ] =====

const (
	// AppName - 어플리케이션 명
	AppName = "cb-restapigw"
	// AppVersion - 어플리케이션 버전
	AppVersion = "0.1.0"
	// AppHeaderName - 어플리케이션 식별을 위한 Header 관리 명
	AppHeaderName = "X-CB-RESTAPIGW"
	// AppUserAgent - Backend 전달에 사용할 User Agent Header 값
	AppUserAgent = AppName + " version " + AppVersion
)

// ===== [ Types ] =====

// ===== [ Implementations ] =====

// ===== [ Private Functions ] =====

// ===== [ Public Functions ] =====

// GetStrings - 지정된 맵 데이터에서 지정된 이름에 해당하는 데이터를 []string 으로 반환
func GetStrings(data map[string]interface{}, name string) []string {
	result := []string{}
	if datas, ok := data[name]; ok {
		if data, ok := datas.([]interface{}); ok {
			for _, val := range data {
				if strVal, ok := val.(string); ok {
					result = append(result, strVal)
				}
			}
		}
	}
	return result
}

// GetString - 지정된 맵 데이터에서 지정한 키에 해당하는 데이터를 string 으로 반환
func GetString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// GetBool - 지정된 맵 데이터에서 지정한 키에 해당하는 데이터를 bool 으로 반환
func GetBool(data map[string]interface{}, key string) bool {
	if val, ok := data[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// GetInt64 - 지정된 맵 데이터에서 지정한 키에 해당하는 데이터를 int64 로 반환
func GetInt64(data map[string]interface{}, key string) int64 {
	if val, ok := data[key]; ok {
		switch i := val.(type) {
		case int64:
			return i
		case int:
			return int64(i)
		case float64:
			return int64(i)
		}
	}
	return 0
}

// ContainsString returns true if a string is present in a iteratee.
func ContainsString(s []string, v string) bool {
	for _, vv := range s {
		if vv == v {
			return true
		}
	}
	return false
}

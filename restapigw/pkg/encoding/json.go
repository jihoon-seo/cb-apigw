package encoding

import (
	"encoding/json"
	"io"
)

// ===== [ Constants and Variables ] =====

const (
	// JSON - JSON 인코딩 식별자
	JSON = "json"
)

// ===== [ Types ] =====

// ===== [ Implementations ] =====

// ===== [ Private Functions ] =====

// ===== [ Public Functions ] =====

// JSONDecoder - 지정한 Reader의 JSON 데이터에 대한 Decoder를 생성하고 Decode 처리
func JSONDecoder(r io.Reader, v *map[string]interface{}) error {
	d := json.NewDecoder(r)
	d.UseNumber()
	return d.Decode(v)
}

// JSONCollectionDecoder - 지정한 Reader의 JSON 데이터에 대한 Collection 으로 Decoder를 생성하고 Decode 처리
func JSONCollectionDecoder(r io.Reader, v *map[string]interface{}) error {
	var collection []interface{}
	d := json.NewDecoder(r)
	d.UseNumber()
	if err := d.Decode(&collection); err != nil {
		return err
	}
	// Backend 결과 Array를 처리하기 위한 식별자 설정
	*(v) = map[string]interface{}{"collection": collection}
	return nil
}

// NewJSONDecoder - Collection 여부에 따라서 JSON Decoder 생성
func NewJSONDecoder(isCollection bool) func(io.Reader, *map[string]interface{}) error {
	if isCollection {
		return JSONCollectionDecoder
	}
	return JSONDecoder
}

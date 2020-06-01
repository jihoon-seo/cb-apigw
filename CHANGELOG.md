### v0.1.9
1. API call 의 Query param 을 forward 하지 않는 오류 수정 ([0dc7753](https://github.com/cloud-barista/cb-apigw/commit/0dc775362cd5011adf851d598f83a10763b70f32))
2. JSON array 관련 ([e62b3c1](https://github.com/cloud-barista/cb-apigw/commit/e62b3c19b8ee9051573f376601d893dba455fa92))
    1.  `is_collection: true` 인 경우
        1. `wrap_collection_to_json: true` 인 경우: 응답을 `"collection"` 이라는 필드의 객체 형식으로 반환
        2. `wrap_collection_to_json: false` 인 경우: 응답을 Array 형태로 반환
3. Bypass 기능 추가 ([4573e84](https://github.com/cloud-barista/cb-apigw/commit/4573e8492a7fa22026fb6be4183cdc770eb80778))
4. Query param, HTTP header 의 전달 정책을 whitelist 에서 blacklist 로 변경
5. Rate Limit 기능 추가 ([1d9911b](https://github.com/cloud-barista/cb-apigw/commit/1d9911ba83057e3d708fba0731f2d33aec555729))

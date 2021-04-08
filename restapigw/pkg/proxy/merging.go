// Package proxy - Backend의 결과들을 Merge 처리하는 Merging 패키지
package proxy

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/errors"
)

// ===== [ Constants and Variables ] =====

const (
	defaultCombinerName = "default"
	sequentialKey       = "sequential"
)

var (
	errNullResult     = errors.New("invalid response")
	responseCombiners = initResponseCombiners()
	reMergeKey        = regexp.MustCompile(`\{\{\.Resp(\d+)_([\d\w-_\.]+)\}\}`)
)

// ===== [ Types ] =====

type (
	// incrementalMergeAccumulator - 점진적인 Merging 처리를 위한 데이터 구조
	incrementalMergeAccumulator struct {
		pending  int
		data     *Response
		combiner ResponseCombiner
		errs     []error
	}

	// mergeError - Merging 과정에서 발생하는 오류들 관리 구조
	mergeError struct {
		errs []error
	}

	// ResponseCombiner - 여러 Response의 데이터를 Merging 처리해서 하나의 Response 데이터로 구성하는 함수 정의
	ResponseCombiner func(int, []*Response) *Response

	// partResult - Backend 처리를 통해 반환된 http.Response와 발생한 Error 정보를 관리하는 구조 정의
	partResult struct {
		Response *Response
		Error    error
	}
)

// ===== [ Implementations ] =====

// Merge - 지정한 Response에 대한 점진적인 Merging 처리
func (ima *incrementalMergeAccumulator) Merge(res *Response, err error) {
	ima.pending--
	if err != nil {
		ima.errs = append(ima.errs, err)
		if ima.data != nil {
			ima.data.IsComplete = false
		} else if res != nil {
			// Error 상태지만 Response가 존재하는 경우
			ima.data = res
			ima.data.IsComplete = false
		}
		return
	}
	if res == nil {
		// 정상이지만 Response가 없는 경우
		ima.errs = append(ima.errs, errNullResult)
		return
	}

	if ima.data == nil {
		// 정상이면 Resposne 존재하는 경우
		ima.data = res
		return
	}

	// 이전 데이터와 Merge 처리
	ima.data = ima.combiner(2, []*Response{ima.data, res})
}

// Result - 처리된 Merging 결과 반환
func (ima *incrementalMergeAccumulator) Result() (*Response, error) {
	if ima.data == nil {
		return &Response{Data: make(map[string]interface{}), IsComplete: false}, newMergeError(ima.errs)
	}

	if ima.pending != 0 || len(ima.errs) != 0 {
		ima.data.IsComplete = false
	}
	return ima.data, newMergeError(ima.errs)
}

// Error - Merging 작업 중에 발생한 오류 메시지 반환
func (me mergeError) Error() string {
	msg := make([]string, len(me.errs))
	for i, err := range me.errs {
		msg[i] = err.Error()
	}
	return strings.Join(msg, "\n")
}

// ===== [ Private Functions ] =====

// newMergeError - Merging 처리 중에 발생한 오류들을 하나의 오류로 반환
func newMergeError(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	return mergeError{errs}
}

// requestPart - 지정한 요청을 호출하고 오류와 Response 정보를 반환
func requestPart(ctx context.Context, next Proxy, req *Request, out chan<- *partResult) {
	localCtx, cancel := context.WithCancel(ctx)

	// Backend Request 호출
	res, err := next(localCtx, req)
	pr := &partResult{}

	// 오류가 발생한 경우
	if err != nil {
		pr.Error = err
		if res == nil {
			// 오류 발생
			out <- pr
			cancel()
			return
		}

		// 오류가 발생했지만 Reponse가 존재하는 경우
		pr.Response = res
		out <- pr
		cancel()
		return
	} else if res == nil {
		// 오류없이 Empty Response인 경우
		pr.Error = errNullResult
		out <- pr
		cancel()
		return
	} else {
		// 정상 처리
		pr.Response = res
	}

	select {
	case out <- pr:
	case <-ctx.Done():
		pr.Error = ctx.Err()
		out <- pr
	}

	cancel()
}

// newIncrementalMergeAccumultor - 지정한 Backend count 와 ResponseCombiner를 설정한 점진적인 Merge 처리기 생성
func newIncrementalMergeAccumultor(backendCount int, rc ResponseCombiner) *incrementalMergeAccumulator {
	return &incrementalMergeAccumulator{
		pending:  backendCount,
		combiner: rc,
		errs:     []error{},
	}
}

// combineData - 지정한 Backend count와 Response들을 기준으로 Merging 처리된 Response 반환
func combineData(backendCount int, reses []*Response) *Response {
	isComplete := len(reses) == backendCount
	var mergedResponse *Response
	for _, res := range reses {
		if res == nil || res.Data == nil {
			isComplete = false
			continue
		}

		isComplete = isComplete && res.IsComplete
		if mergedResponse == nil {
			mergedResponse = res
			continue
		}

		for k, v := range res.Data {
			mergedResponse.Data[k] = v
		}
	}

	if mergedResponse == nil {
		// do not allow nil data to response
		return &Response{Data: make(map[string]interface{}), IsComplete: isComplete}
	}
	mergedResponse.IsComplete = isComplete
	return mergedResponse
}

// initResponseCombiners - Response 데이터를 Merging 하는 ResponseCombiner 초기화
func initResponseCombiners() *combinerRegister {
	return newCombinerRegister(map[string]ResponseCombiner{defaultCombinerName: combineData}, combineData)
}

// getResponseCombiner - 기본적으로 사용되는 ResponseCombiner 반환
func getResponseCombiner() ResponseCombiner {
	combiner, _ := responseCombiners.GetResponseCombiner(defaultCombinerName)
	return combiner
}

// shouldRunSequentialMerger - 지정된 설정 정보를 기준으로 Merging이 순차 처리가 되어야할지 검증
func shouldRunSequentialMerger(eConf *config.EndpointConfig) bool {
	if v, ok := eConf.Middleware[MWNamespace]; ok {
		if e, ok := v.(config.MWConfig); ok {
			if v, ok := e[sequentialKey]; ok {
				c, ok := v.(bool)
				return ok && c
			}
		}
	}
	return false
}

// parallelMerge - 지정한 시간내에 Timeout 발생하는 Context 기반으로 Request를 처리하고 도착하는 Response를 병렬로 처리
func parallelMerge(timeout time.Duration, rc ResponseCombiner, next ...Proxy) Proxy {
	return func(ctx context.Context, req *Request) (*Response, error) {
		localCtx, cancel := context.WithTimeout(ctx, timeout)

		out := make(chan *partResult, len(next))

		// 병렬로 Backend 호출
		for _, n := range next {
			go requestPart(localCtx, n, req, out)
		}

		acc := newIncrementalMergeAccumultor(len(next), rc)
		for i := 0; i < len(next); i++ {
			resultData := <-out
			acc.Merge(resultData.Response, resultData.Error)
		}

		result, err := acc.Result()
		cancel()
		return result, err
	}
}

// sequentialMerge - 지정한 시간내에 Timeout 발생하는 Context 기반으로 Request를 순차적으로 처리하고 이전 Response의 결과를 파라미터로 처리해서 다음 Request를 처리하는 방식으로 순차 처리
func sequentialMerge(patterns []string, timeout time.Duration, rc ResponseCombiner, next ...Proxy) Proxy {
	return func(ctx context.Context, req *Request) (*Response, error) {
		localCtx, cancel := context.WithTimeout(ctx, timeout)

		parts := make([]*Response, len(next))

		out := make(chan *partResult, 1)

		acc := newIncrementalMergeAccumultor(len(next), rc)
	TxLoop:
		for i, n := range next {
			// 두번째 부터 전 호출의 결과에서 파라미터 검증
			if i > 0 {
				for _, match := range reMergeKey.FindAllStringSubmatch(patterns[i], -1) {
					if len(match) > 1 {
						rNum, err := strconv.Atoi(match[1])
						if err != nil || rNum >= i || parts[rNum] == nil {
							continue
						}
						key := "Resp" + match[1] + "_" + match[2]

						var v interface{}
						var ok bool

						data := parts[rNum].Data
						keys := strings.Split(match[2], ".")
						if len(keys) > 1 {
							for _, k := range keys[:len(keys)-1] {
								v, ok = data[k]
								if !ok {
									break
								}
								switch clean := v.(type) {
								case map[string]interface{}:
									data = clean
								default:
									break

								}
							}
						}

						v, ok = data[keys[len(keys)-1]]
						if !ok {
							continue
						}
						switch clean := v.(type) {
						case string:
							req.Params[key] = clean
						case int:
							req.Params[key] = strconv.Itoa(clean)
						case float64:
							req.Params[key] = strconv.FormatFloat(clean, 'E', -1, 32)
						case bool:
							req.Params[key] = strconv.FormatBool(clean)
						default:
							req.Params[key] = fmt.Sprintf("%v", v)
						}
					}
				}
			}

			// 순차적 호출
			requestPart(localCtx, n, req, out)
			resultData := <-out
			if resultData.Error != nil {
				if i == 0 {
					cancel()
					return resultData.Response, resultData.Error
				}

				acc.Merge(resultData.Response, resultData.Error)
				break TxLoop
			} else {
				acc.Merge(resultData.Response, resultData.Error)
				if resultData.Response.IsComplete {
					break TxLoop
				}
				parts[i] = resultData.Response
			}
		}

		result, err := acc.Result()
		cancel()
		return result, err
	}
}

// ===== [ Public Functions ] =====

// NewMergeDataChain - 전달된 Endpoint 설정을 기준으로 Backend 갯수에 따라서 Response를 Merging 하는 Proxy Call chain 생성
func NewMergeDataChain(eConf *config.EndpointConfig) CallChain {
	totalBackends := len(eConf.Backend)
	if totalBackends == 0 {
		panic(ErrNoBackends)
	}
	if totalBackends == 1 {
		return EmptyChain
	}

	serviceTimeout := time.Duration(85*eConf.Timeout.Nanoseconds()/100) * time.Nanosecond
	combiner := getResponseCombiner()

	return func(next ...Proxy) Proxy {
		if len(next) != totalBackends {
			panic(ErrNotEnoughProxies)
		}
		if !shouldRunSequentialMerger(eConf) {
			return parallelMerge(serviceTimeout, combiner, next...)
		}
		patterns := make([]string, totalBackends)
		for i, b := range eConf.Backend {
			patterns[i] = b.URLPattern
		}
		return sequentialMerge(patterns, serviceTimeout, combiner, next...)
	}
}

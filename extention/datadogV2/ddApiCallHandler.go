package extV2

import (
	_fmt "fmt"
	_nethttp "net/http"
	_os "os"
	_reflect "reflect"
	_runtime "runtime"
)

// APICall represents an API call with metadata
type APICall struct {
	MethodName string
	APIName    string
}

// CallWithErrorHandling executes an API call and handles errors
func (ac *APICall) CallWithErrorHandling(fn func() (interface{}, *_nethttp.Response, error)) (interface{}, *_nethttp.Response, error) {
	resp, r, err := fn()
	if err != nil {
		_fmt.Fprintf(_os.Stderr, "Error when calling `%s.%s`: %v\n", ac.APIName, ac.MethodName, err)
		_fmt.Fprintf(_os.Stderr, "Full HTTP response: %v\n", r)
	}
	return resp, r, err
}

// NewAPICall creates a new APICall with automatic method name detection
func NewAPICall(apiName string, fn interface{}) *APICall {
	methodName := _runtime.FuncForPC(_reflect.ValueOf(fn).Pointer()).Name()
	// Extract just the method name from the full path
	if lastDot := len(methodName) - 1; lastDot >= 0 {
		for i := lastDot; i >= 0; i-- {
			if methodName[i] == '.' {
				methodName = methodName[i+1:]
				break
			}
		}
	}

	return &APICall{
		MethodName: methodName,
		APIName:    apiName,
	}
}

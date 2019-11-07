package serve

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"gitlab.jiangxingai.com/applications/base-modules/internal-sdk/go-utils/logger"
)

const mimeJSON = "application/json; charset=utf-8"

// APIError api 异常类型
type APIError struct {
	Code        int    `json:"code"`
	Description string `json:"desc"`
}

// Error 服务报错
func Error(w http.ResponseWriter, reason string, code int) {
	w.Header().Set("Content-Type", mimeJSON)

	e := APIError{code, reason}

	w.WriteHeader(code)

	rs, err := json.Marshal(e)
	if err != nil {
		logger.Fatal(err)
	}

	w.Write(rs)
}

// ErrorWithCode 使用预设的 Http 状态码抛出异常
func ErrorWithCode(w http.ResponseWriter, code int) {
	Error(w, http.StatusText(code), code)
}

// ErrorNotFound  服务报错
func ErrorNotFound(w http.ResponseWriter) {
	ErrorWithCode(w, http.StatusNotFound)
}

type notFoundHandler struct {
}

func (h *notFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger.Infof("In: 404 \t%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

	ErrorNotFound(w)
}

// NewNotFoundHandler 404 处理问题
func NewNotFoundHandler() http.Handler {
	return &notFoundHandler{}
}

// APIResult API 返回结果
type APIResult struct {
	Data        *map[string]interface{} `json:"data,omitempty"`
	Description string                  `json:"desc"`
}

// NewAPIResult 使用对象
func NewAPIResult(data *map[string]interface{}) *APIResult {

	return &APIResult{Data: data, Description: "success"}
}

// WriteData 将结果写入相应
func WriteData(w http.ResponseWriter, data *map[string]interface{}) {
	WriteResult(w, NewAPIResult(data))
}

// WriteResult 返回结果
func WriteResult(w http.ResponseWriter, result *APIResult) {
	w.Header().Set("Content-Type", mimeJSON)
	w.WriteHeader(http.StatusOK)

	rs, err := json.Marshal(result)
	if err != nil {
		logger.Fatal(err)
	}

	w.Write(rs)
}

// WriteSucess 写入标记操作为成功的空响应
func WriteSucess(w http.ResponseWriter) {
	WriteData(w, nil)
}

// SimpleLoggingMw 简单的日志中间件
func SimpleLoggingMw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Infof("In: \t%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		start := time.Now()
		next.ServeHTTP(w, r)
		cost := time.Since(start)

		resp := r.Response
		if resp != nil {
			logger.Infof("End:\t%s %s %s %dms %s\n", r.RemoteAddr, r.Method, r.URL, cost.Nanoseconds()/int64(time.Millisecond), resp.Status)
		} else {
			logger.Infof("End:\t%s %s %s %dms\n", r.RemoteAddr, r.Method, r.URL, cost.Nanoseconds()/int64(time.Millisecond))
		}

	})
}

// RecoverMiddleware panic 发生时，返回结果
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				var err error
				switch x := r.(type) {
				case string:
					err = errors.New(x)
				case error:
					err = x
				default:
					err = errors.New("Unknown panic")
				}
				logger.Error(err)
				Error(w, err.Error(), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

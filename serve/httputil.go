package serve

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
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
		log.Fatalln(err)
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
	log.Printf("In: 404 \t%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

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
		log.Fatalln(err)
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
		log.Printf("In:\t%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		start := time.Now()
		next.ServeHTTP(w, r)
		cost := time.Since(start)
		log.Printf("End:\t%s %s %s %dms\n", r.RemoteAddr, r.Method, r.URL, cost.Nanoseconds()/int64(time.Millisecond))
	})
}

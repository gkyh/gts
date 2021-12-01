package gts

import (
	"encoding/json"
	"net/http"
)

type Result struct {
	Message string      `json:"msg"`
	Code    int         `json:"code"` // 200 means success, other means fail
	Data    interface{} `json:"data"`
}

const StatusOk int = 200       // success
const StatusUnknown int = 500  // fail,reason is unknown
const StatusNotFound int = 404 // fail,reason is not found

const BadReq int = 400    //提交错误
const NotAuthor int = 401 //认证失败或操作超时
const Repeat int = 402    //请求处理已完成或重复请求
const NotPermis int = 403 //没有操作权限

const Forbidden int = 405   //访问被禁止
const StatusError int = 406 //请求业务状态错误
const Failed int = 410      //请求业务无法成功受理，业务自定义

const Exception int = 500  //服务器异常，程序抛出异常
const NotLimit int = 501   //不接受请求，如IP被限制
const BadGateway int = 502 //外部异常，通常为请求第三方
const Unavilable int = 503 //请求未能应答

type ResultBuilder struct {
	result *Result
	writer http.ResponseWriter
}

func NewResp(w http.ResponseWriter) ResultBuilder {
	return ResultBuilder{result: &Result{Code: StatusOk, Message: "", Data: nil}, writer: w}
}

// success default is true, code default is 200
func NewResult() ResultBuilder {
	return ResultBuilder{result: &Result{Code: StatusOk, Message: "", Data: nil}}
}

func (builder ResultBuilder) Code(code int) ResultBuilder {
	builder.result.Code = code
	return builder
}
func (builder ResultBuilder) Error(code int) ResultBuilder {
	builder.result.Code = code
	return builder
}

func (builder ResultBuilder) Message(message string) ResultBuilder {
	builder.result.Message = message
	return builder
}
func (builder ResultBuilder) NotPermis() ResultBuilder {

	builder.result.Code = NotPermis
	builder.result.Message = "没有操作权限，se000401"
	return builder
}
func (builder ResultBuilder) NotFound() ResultBuilder {

	builder.result.Code = StatusNotFound
	builder.result.Message = "未找到的记录，re000404"
	return builder
}

func (builder ResultBuilder) Success() ResultBuilder {

	builder.result.Code = StatusOk
	builder.result.Message = "OK"
	return builder
}

func (builder ResultBuilder) OK() ResultBuilder {

	builder.result.Code = StatusOk
	return builder
}

func (builder ResultBuilder) Fail() ResultBuilder {

	builder.result.Code = StatusUnknown
	return builder
}

func (builder ResultBuilder) Data(data interface{}) ResultBuilder {
	builder.result.Data = data
	return builder
}

func (builder ResultBuilder) Build() *Result {
	return builder.result
}

func (builder ResultBuilder) Bd() {

	w := builder.writer
	w.Header().Set("Content-Type", "application/Json; charset=utf-8")
	json.NewEncoder(w).Encode(builder.result)
}

type Resource struct {
	Message     string      `json:"msg"`
	Code        int         `json:"code"` // 200 means success, other means fail
	Data        interface{} `json:"data"`
	TotalPage   int32       `json:"totalPage"`
	PageSize    string      `json:"pageSize"`
	CurrentPage string      `json:"currentPage"`
}

type ResourceBuilder struct {
	resource *Resource
	writer   http.ResponseWriter
}

func NewResData(w http.ResponseWriter) ResourceBuilder {
	return ResourceBuilder{resource: &Resource{Code: StatusOk, Message: "OK", Data: nil, TotalPage: 0, PageSize: "0", CurrentPage: "0"}, writer: w}
}

// success default is true, code default is 200
func NewResource(data interface{}) ResourceBuilder {
	return ResourceBuilder{resource: &Resource{Code: StatusOk, Message: "OK", Data: data, TotalPage: 0, PageSize: "0", CurrentPage: "0"}}
}

func (builder ResourceBuilder) Code(code int) ResourceBuilder {
	builder.resource.Code = code
	return builder
}
func (builder ResourceBuilder) Message(s string) ResourceBuilder {
	builder.resource.Message = s
	return builder
}

func (builder ResourceBuilder) Data(data interface{}) ResourceBuilder {
	builder.resource.Data = data
	return builder
}

func (builder ResourceBuilder) TotalPage(i int32) ResourceBuilder {
	builder.resource.TotalPage = i
	return builder
}

func (builder ResourceBuilder) PageSize(i string) ResourceBuilder {
	builder.resource.PageSize = i
	return builder
}

func (builder ResourceBuilder) CurrentPage(i string) ResourceBuilder {
	builder.resource.CurrentPage = i
	return builder
}

func (builder ResourceBuilder) Build() *Resource {
	return builder.resource
}

func (builder ResourceBuilder) Bd() {

	w := builder.writer
	w.Header().Set("Content-Type", "application/Json; charset=utf-8")
	json.NewEncoder(w).Encode(builder.resource)
}

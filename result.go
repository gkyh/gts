package gts

import (
	"encoding/json"
	"net/http"
)

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

type Result struct {
	Message string      `json:"msg"`
	Code    int         `json:"code"` // 200 means success, other means fail
	Data    interface{} `json:"data"`
}

type Resource struct {
	Message     string      `json:"msg"`
	Code        int         `json:"code"`
	Data        interface{} `json:"data"`
	Total       int32       `json:"total"`
	TotalPage   int32       `json:"totalPage"`
	PageSize    string      `json:"pageSize"`
	CurrentPage string      `json:"currentPage"`
}

// ---- Builder ----
type RespBuilder struct {
	writer http.ResponseWriter
	isPage bool
	result *Result
	res    *Resource
}

type ResultBuilder struct {
	result *Result
	writer http.ResponseWriter
}

func NewResp(w http.ResponseWriter) *RespBuilder {
	return &RespBuilder{
		writer: w,
		result: &Result{Code: StatusOk, Message: "OK"},
	}
}
// ---- 通用方法 ----
func (b *RespBuilder) Code(code int) *RespBuilder {
	if b.isPage {
		b.res.Code = code
	} else {
		b.result.Code = code
	}
	return b
}

func (b *RespBuilder) Message(msg string) *RespBuilder {
	if b.isPage {
		b.res.Message = msg
	} else {
		b.result.Message = msg
	}
	return b
}

func (b *RespBuilder) Data(data interface{}) *RespBuilder {
	if b.isPage {
		b.res.Data = data
	} else {
		b.result.Data = data
	}
	return b
}

// ---- 分页相关方法 ----
func (b *RespBuilder) ensurePage() {
	if !b.isPage {
		b.isPage = true
		b.res = &Resource{
			Code:    b.result.Code,
			Message: b.result.Message,
			Data:    b.result.Data,
		}
		b.result = nil
	}
}

func (b *RespBuilder) Total(total int32) *RespBuilder {
	b.ensurePage()
	b.res.Total = total
	return b
}

func (b *RespBuilder) TotalPage(totalPage int32) *RespBuilder {
	b.ensurePage()
	b.res.TotalPage = totalPage
	return b
}

func (b *RespBuilder) PageSize(size string) *RespBuilder {
	b.ensurePage()
	b.res.PageSize = size
	return b
}

func (b *RespBuilder) CurrentPage(page string) *RespBuilder {
	b.ensurePage()
	b.res.CurrentPage = page
	return b
}

// ---- 输出 ----
func (b *RespBuilder) JSON() {
	w := b.writer
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if b.isPage {
		_ = json.NewEncoder(w).Encode(b.res)
	} else {
		_ = json.NewEncoder(w).Encode(b.result)
	}
}

func (b *RespBuilder) Error(code int) *RespBuilder {
	return b.Code(code)
}

// 没有操作权限
func (b *RespBuilder) NotPermis() *RespBuilder {
	return b.Code(NotPermis).Message("没有操作权限，se000401")
}

// 未找到
func (b *RespBuilder) NotFound() *RespBuilder {
	return b.Code(StatusNotFound).Message("未找到的记录，re000404")
}

// 成功
func (b *RespBuilder) Success() *RespBuilder {
	return b.Code(StatusOk).Message("OK")
}

// OK（只设置状态码）
func (b *RespBuilder) OK() *RespBuilder {
	return b.Code(StatusOk)
}

// 失败（未知原因）
func (b *RespBuilder) Fail() *RespBuilder {
	return b.Code(StatusUnknown)
}


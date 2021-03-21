package gts

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type M map[string]interface{}
type Context struct {
	Writer   http.ResponseWriter
	Request  *http.Request
	Sessions Session
}

func (c *Context) ReqValue(params ...string) map[string]interface{} {

	//req.ParseForm()
	req := c.Request
	m := make(map[string]interface{})
	for _, value := range params {

		val := req.FormValue(value)
		if val == "null" || val == "" {
			continue
		}
		m[value] = req.FormValue(value)
	}
	return m
}

func (c *Context) FormValue(key, val string) string {

	str := c.Request.FormValue(key)

	if str == "" {
		return val
	} else {
		return str
	}
}

func (c *Context) Session() *Store {

	sid, _ := c.Sessions.SessionID(c.Request)
	return &Store{
		SessionID: sid,
		Sessions:  c.Sessions,
		Response:  c.Writer,
	}
}
func (c *Context) Write(status int, b []byte) {

	print(string(b))
	c.Writer.WriteHeader(status)
	c.Writer.Write(b)
}

func (c *Context) WriteString(s string) {

	print(s)
	c.Writer.WriteHeader(200)
	io.WriteString(c.Writer, s)
}

func (c *Context) HTML(status int, s string) {

	print(s)
	w := c.Writer
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	w.Write([]byte(s))
}

func (c *Context) JSON(status int, m map[string]interface{}) {

	w := c.Writer
	w.Header().Set("Content-Type", "application/Json; charset=utf-8")
	w.WriteHeader(status)
	jsonBytes, _ := json.Marshal(m)
	print(string(jsonBytes))
	w.Write(jsonBytes)
}
func (c *Context) Map(m map[string]interface{}) {

	w := c.Writer
	w.Header().Set("Content-Type", "application/Json; charset=utf-8")
	print(m)
	json.NewEncoder(w).Encode(m)
}

func (c *Context) Result(s string) {

	w := c.Writer
	w.Header().Set("Content-Type", "application/Json; charset=utf-8")
	print(s)
	io.WriteString(w, s)
}
func (c *Context) Err(code int32, s string) {

	w := c.Writer
	w.Header().Set("Content-Type", "application/Json; charset=utf-8")
	str := fmt.Sprintf(`{"code": %d, "msg": "%s"}`, code, s)
	print(str)
	io.WriteString(w, str)
}

func (c *Context) Msg(s string) {

	w := c.Writer
	w.Header().Set("Content-Type", "application/Json; charset=utf-8")
	str := fmt.Sprintf(`{"code": 200, "msg": "%s"}`, s)
	print(str)
	io.WriteString(w, str)
}

func (c *Context) OK() {

	w := c.Writer
	w.Header().Set("Content-Type", "application/Json; charset=utf-8")
	//w.WriteHeader(http.StatusOK)
	str := `{"code": 200, "msg": "处理成功"}`
	print(str)
	io.WriteString(w, str)
}

func (c *Context) Redirect(url string) {

	w := c.Writer
	r := c.Request
	print("Redirect:" + url)
	http.Redirect(w, r, url, http.StatusFound)
}

func (c *Context) SessionID() (string, bool) {

	session := c.Sessions
	return session.SessionID(c.Request)
}

func (c *Context) Set(key string, v map[string]interface{}) {

	ctx := context.WithValue(c.Request.Context(), key, v)
	c.Request = c.Request.WithContext(ctx)
}

func (c *Context) Get(key string) map[string]interface{} {

	return c.Request.Context().Value(key).(map[string]interface{})
}
func (c *Context) SetString(key, value string) {

	ctx := context.WithValue(c.Request.Context(), key, value)
	c.Request = c.Request.WithContext(ctx)
}

func (c *Context) GetString(key string) string {

	v := c.Request.Context().Value(key)
	if v != nil {
		return v.(string)
	}
	return ""
}

package gts

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

type M map[string]interface{}
type Context struct {
	Writer   http.ResponseWriter
	Request  *http.Request
	Sessions *Session
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

	c.Writer.WriteHeader(status)
	c.Writer.Write(b)
}

func (c *Context) WriteString(s string) {

	//c.Writer.WriteHeader(status)
	io.WriteString(c.Writer, s)
}

func (c *Context) HTML(status int, s string) {

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
	w.Write(jsonBytes)
}
func (c *Context) Map(m map[string]interface{}) {

	w := c.Writer
	w.Header().Set("Content-Type", "application/Json; charset=utf-8")
	//w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(m)
}

func (c *Context) Result(s string) {

	w := c.Writer
	w.Header().Set("Content-Type", "application/Json; charset=utf-8")
	//w.WriteHeader(http.StatusOK)
	io.WriteString(w, s)
}

func (c *Context) Msg(s string) {

	w := c.Writer
	w.Header().Set("Content-Type", "application/Json; charset=utf-8")
	//w.WriteHeader(http.StatusOK)
	io.WriteString(w, `{"code": 200, "msg": "`+s+`"}`)
}

func (c *Context) OK() {

	w := c.Writer
	w.Header().Set("Content-Type", "application/Json; charset=utf-8")
	//w.WriteHeader(http.StatusOK)
	io.WriteString(w, `{"code": 200, "msg": "处理成功"}`)
}

func (c *Context) Redirect(url string) {

	w := c.Writer
	r := c.Request
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

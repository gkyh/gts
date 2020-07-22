package gts

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

type Context struct {
	Writer   http.ResponseWriter
	Request  *http.Request
	Sessions *Session
}

func (c *Context) Write(b []byte) {

	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write(b)
}

func (c *Context) WriteString(status int, s string) {

	c.Writer.WriteHeader(status)
	io.WriteString(c.Writer, s)
}

func (c *Context) HTML(status int, s string) {

	w := c.Writer
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	io.WriteString(w, s)
}

func (c *Context) JSON(status int, m map[string]interface{}) {

	w := c.Writer
	w.Header().Set("Content-Type", "application/Json; charset=utf-8")
	w.WriteHeader(status)
	jsonBytes, _ := json.Marshal(m)
	w.Write(jsonBytes)
}

func (c *Context) Result(s string) {

	w := c.Writer
	w.Header().Set("Content-Type", "application/Json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, s)
}

func (c *Context) Msg(s string) {

	w := c.Writer
	w.Header().Set("Content-Type", "application/Json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, `{"code": 200, "msg": "`+s+`"}`)
}

func (c *Context) OK() {

	w := c.Writer
	w.Header().Set("Content-Type", "application/Json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
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
func (c *Context) SessionVal(key interface{}) (interface{}, bool) {

	session := c.Sessions
	if session == nil {
		return nil, false
	}
	return session.Get(c.Request, key)
}
func (c *Context) GetSession(key interface{}) interface{} {

	session := c.Sessions
	if session == nil {
		return nil
	}
	i, b := session.Get(c.Request, key)
	if b {
		return i
	} else {

		return nil
	}
}
func (c *Context) SetSession(key interface{}, value interface{}) bool {

	w := c.Writer
	r := c.Request

	session := c.Sessions
	if session == nil {
		return false
	}

	sid, _ := session.SessionID(r)
	if sid == "" {

		sid = session.New(w)

		session.SetVal(sid, key, value)
		return true
	}
	return session.Set(r, key, value)

}
func (c *Context) Set(key string, v map[string]interface{}) {

	ctx := context.WithValue(c.Request.Context(), key, v)
	c.Request = c.Request.WithContext(ctx)
}

func (c *Context) Get(key string) map[string]interface{} {

	return c.Request.Context().Value(key).(map[string]interface{})
}

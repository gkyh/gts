package gts

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

type Router struct {
	rLen   []int
	routes []map[string]http.HandlerFunc
	mws    []HandlerFun
	ses    *Session
}

var (
	fileLen    = 0
	fileRoutes map[string]http.HandlerFunc
	Type       = map[string]int{
		"Any":    0,
		"GET":    1,
		"POST":   2,
		"DELETE": 3,
		"PUT":    4,
	}
)

func New() *Router {

	fileRoutes = make(map[string]http.HandlerFunc)

	return &Router{routes: []map[string]http.HandlerFunc{fileRoutes, fileRoutes, fileRoutes, fileRoutes, fileRoutes}, rLen: make([]int, 5), ses: nil}
}

func (p *Router) Cookie(cookieName string, maxLifeTime, cookieTime int64) {

	p.ses = NewSession(cookieName, maxLifeTime, cookieTime)
}

type HandlerFun func(http.HandlerFunc) http.HandlerFunc

func (p *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	method := r.Method

	fmt.Println("[" + method + "]" + url)
	if p.rLen[0] > 0 {
		if fun, ok := p.routes[0][url]; ok {

			if p.ses == nil {
				fun(w, r)
			} else {
				fun(w, r.WithContext(context.WithValue(r.Context(), "session", p.ses)))
			}
			return
		}
	}

	var t int = 0
	t = Type[method]
	if t > 0 && p.rLen[t] > 0 {

		if fun, ok := p.routes[t][url]; ok {
			if p.ses == nil {
				fun(w, r)
			} else {
				fun(w, r.WithContext(context.WithValue(r.Context(), "session", p.ses)))
			}
			return
		}
	}

	//静态资源
	if fileLen > 0 {
		for k, f := range fileRoutes {

			if strings.HasPrefix(url, k) {

				f(w, r)
				return
			}
		}
	}
	http.Error(w, "error URL:"+r.URL.String(), http.StatusBadRequest)

}

func middleware(mws []HandlerFun, h http.HandlerFunc) http.HandlerFunc {

	l := len(mws)
	for i := l - 1; i >= 0; i-- {

		h = run(mws[i], h)
	}
	return h
}

func run(m HandlerFun, h http.HandlerFunc) http.HandlerFunc {

	return m(h)
}

func (p *Router) R(i int, path string, h http.HandlerFunc, f ...HandlerFun) {

	m := p.routes[i]
	if len(f) > 0 {

		m[path] = middleware(p.mws, f[0](h))
	} else {

		m[path] = middleware(p.mws, h)
	}
	p.rLen[i]++
}

func (p *Router) Static(relativePath string, dirPath string) {

	fileRoutes[relativePath] = func(w http.ResponseWriter, r *http.Request) {

		http.StripPrefix(relativePath, http.FileServer(http.Dir(dirPath))).ServeHTTP(w, r)
	}
	fileLen++
}
func (p *Router) Use(h HandlerFun) {

	p.mws = append(p.mws, h)
}

func (p *Router) Any(relativePath string, handler http.HandlerFunc, filter ...HandlerFun) {

	p.R(0, relativePath, handler, filter...)

}
func (p *Router) Get(relativePath string, handler http.HandlerFunc, filter ...HandlerFun) {

	p.R(1, relativePath, handler, filter...)

}
func (p *Router) Post(relativePath string, handler http.HandlerFunc, filter ...HandlerFun) {

	p.R(2, relativePath, handler, filter...)

}
func (p *Router) Delete(relativePath string, handler http.HandlerFunc, filter ...HandlerFun) {

	p.R(3, relativePath, handler, filter...)

}

func (p *Router) Group(url string, i interface{}, params ...HandlerFun) {

	v := reflect.ValueOf(i)
	t := reflect.TypeOf(i)

	_, ins := t.MethodByName("Router")
	if ins {

		mh := v.MethodByName("Router")
		in := make([]reflect.Value, len(params)+2)
		in[0] = reflect.ValueOf(url)
		in[1] = reflect.ValueOf(p)

		for k, param := range params {
			in[k+2] = reflect.ValueOf(param)
		}
		mh.Call(in)
	}

}

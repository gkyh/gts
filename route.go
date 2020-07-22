package gts

import (
	"fmt"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strings"
)

type IRouter interface {
	Router(*Router)
}

type Router struct {
	rLen   []int
	routes []map[string]HandlerFunc
	mws    []HandlerFun
	ses    *Session
	base   string
}

var (
	fileLen    = 0
	fileRoutes map[string]http.HandlerFunc
	mwRoutes   map[string]HandlerFun
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
	mwRoutes = make(map[string]HandlerFun)

	return &Router{
		routes: []map[string]HandlerFunc{make(map[string]HandlerFunc), make(map[string]HandlerFunc), make(map[string]HandlerFunc), make(map[string]HandlerFunc), make(map[string]HandlerFunc)},
		rLen:   make([]int, 5),
		ses:    nil,
		base:   "",
	}
}

func (p *Router) Cookie(cookieName string, maxLifeTime, cookieTime int64) {

	p.ses = NewSession(cookieName, maxLifeTime, cookieTime)
}

type HandlerFunc func(*http.Request, *Context)
type HandlerFun func(HandlerFunc) HandlerFunc

func (p *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	method := r.Method

	ctx := &Context{Writer: w, Request: r, Sessions: p.ses}

	fmt.Println("[" + method + "]" + url)

	var t int = 0
	t = Type[method]
	if t > 0 && p.rLen[t] > 0 {

		if fun, ok := p.routes[t][url]; ok {
			fun(r, ctx)
			return
		}
	}
	/*
		if p.rLen[0] > 0 {
			if fun, ok := p.routes[0][url]; ok {

				fun(r, ctx)
				return
			}
		}
	*/

	//静态资源
	if fileLen > 0 {
		for k, f := range fileRoutes {

			if strings.HasPrefix(url, k) {

				fmt.Print(k + "\r\n")
				f(w, r)
				return
			}
		}
	}
	http.Error(w, "error URL:"+r.URL.String(), http.StatusBadRequest)

}

func M(mws []HandlerFun, h HandlerFunc) HandlerFunc {

	l := len(mws)
	for i := l - 1; i >= 0; i-- {

		h = run(mws[i], h)
	}

	return F(h)
}

func run(m HandlerFun, h HandlerFunc) HandlerFunc {

	return m(h)
}

func F(h HandlerFunc) HandlerFunc {

	v := reflect.ValueOf(h)
	fn := runtime.FuncForPC(v.Pointer()).Name()

	for k, f := range mwRoutes {

		if strings.Contains(fn, k) {
			return f(h)
		}
	}
	return h
}

func (p *Router) R(i int, path string, h HandlerFunc, f ...HandlerFun) {

	url := p.base + path
	m := p.routes[i]

	vh := reflect.ValueOf(h)
	fn := runtime.FuncForPC(vh.Pointer()).Name()
	fn = fmt.Sprintf("%s:==>:%s\r\n", url, fn)
	fmt.Print(fn)

	if len(f) > 0 {

		m[url] = M(p.mws, f[0](h))
	} else {

		m[url] = M(p.mws, h)
	}
	p.rLen[i]++
}

func (p *Router) Static(relativePath string, dirPath string) {

	fileRoutes[relativePath] = func(w http.ResponseWriter, r *http.Request) {

		http.StripPrefix(relativePath, http.FileServer(http.Dir(dirPath))).ServeHTTP(w, r)
	}
	fileLen++
}
func (p *Router) StaticFs(relativePath string, handler http.HandlerFunc) {

	fileRoutes[relativePath] = handler
	fileLen++
}

func (p *Router) Favicon(dirPath string) {

	fileRoutes["/favicon.ico"] = func(w http.ResponseWriter, r *http.Request) {

		file := dirPath + "favicon.ico"
		if _, err := os.Stat(file); err == nil {
			http.ServeFile(w, r, file)
		}
	}
	fileLen++
}

func (p *Router) Use(h HandlerFun) {

	p.mws = append(p.mws, h)
}

func (p *Router) Any(relativePath string, handler HandlerFunc, filter ...HandlerFun) {

	//p.R(0, relativePath, handler, filter...)
	p.R(1, relativePath, handler, filter...)
	p.R(2, relativePath, handler, filter...)

}

func (p *Router) Get(relativePath string, handler HandlerFunc, filter ...HandlerFun) {

	p.R(1, relativePath, handler, filter...)

}
func (p *Router) Post(relativePath string, handler HandlerFunc, filter ...HandlerFun) {

	p.R(2, relativePath, handler, filter...)

}
func (p *Router) Delete(relativePath string, handler HandlerFunc, filter ...HandlerFun) {

	p.R(3, relativePath, handler, filter...)

}
func (p *Router) Put(relativePath string, handler HandlerFunc, filter ...HandlerFun) {

	p.R(4, relativePath, handler, filter...)

}

func (p *Router) Group(url string, h func(r *Router), params ...HandlerFun) {

	p.base = url
	h(p)
	p.base = ""
}

func (p *Router) Route(url string, i IRouter, params ...HandlerFun) {

	t := reflect.TypeOf(i).String()

	if len(params) > 0 {

		idx := strings.LastIndex(t, ".")
		if idx > 0 {

			idx++
			t = string([]rune(t)[idx:len(t)])
		}
		mwRoutes[t] = params[0]
	}
	p.base = url
	i.Router(p)
	p.base = ""

}

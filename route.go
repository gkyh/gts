package gts

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

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

type HandlerFunc func(*Context)
type HandlerFun func(HandlerFunc) HandlerFunc

func (p *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	method := r.Method

	ctx := &Context{Writer: w, Request: r, Sessions: p.ses}

	fmt.Println("[" + method + "]" + url)
	if p.rLen[0] > 0 {
		if fun, ok := p.routes[0][url]; ok {

			fun(ctx)
			return
		}
	}

	var t int = 0
	t = Type[method]
	if t > 0 && p.rLen[t] > 0 {

		if fun, ok := p.routes[t][url]; ok {
			fun(ctx)
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
	fn = fmt.Sprintf("==>:%s", fn)
	fmt.Print(fn + "\r\n")
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

	fmt.Print(url + ":")
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
func (p *Router) Use(h HandlerFun) {

	p.mws = append(p.mws, h)
}

func (p *Router) Any(relativePath string, handler HandlerFunc, filter ...HandlerFun) {

	p.R(0, relativePath, handler, filter...)

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

func (p *Router) Group(url string, h func(r *Router), params ...HandlerFun) {

	//v := reflect.ValueOf(h)
	//fn := runtime.FuncForPC(v.Pointer()).Name()
	//fn = fmt.Sprintf("G:%s", fn)
	//fmt.Println(fn)
	p.base = url
	h(p)
	p.base = ""
}

func (p *Router) Route(url string, i interface{}, params ...HandlerFun) {

	v := reflect.ValueOf(i)
	t := reflect.TypeOf(i)

	class := fmt.Sprintf("%v", t)

	if len(params) > 0 {

		idx := strings.LastIndex(class, ".")
		if idx > 0 {

			idx++
			class = string([]rune(class)[idx:len(class)])
		}
		mwRoutes[class] = params[0]
	}
	p.base = url
	_, ins := t.MethodByName("Router")
	if ins {

		mh := v.MethodByName("Router")
		in := make([]reflect.Value, 1)
		in[0] = reflect.ValueOf(p)

		//for k, param := range params {
		//	in[k+2] = reflect.ValueOf(param)
		//}
		mh.Call(in)
	}
	p.base = ""

}

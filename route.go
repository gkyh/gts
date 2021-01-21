package gts

import (
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
	rLen    []int
	routes  []map[string]HandlerFunc
	mws     []HandlerFun
	session Session
	base    string
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

var session Session
var logger RouteLogger

type RouteLogger interface {
	Println(v ...interface{})
}

func New() *Router {

	fileRoutes = make(map[string]http.HandlerFunc)
	mwRoutes = make(map[string]HandlerFun)

	return &Router{
		routes:  []map[string]HandlerFunc{make(map[string]HandlerFunc), make(map[string]HandlerFunc), make(map[string]HandlerFunc), make(map[string]HandlerFunc), make(map[string]HandlerFunc)},
		rLen:    make([]int, 5),
		session: nil,
		base:    "",
	}
}

func (p *Router) Cookie(cookieName string, maxLifeTime, cookieTime int64) {

	session = NewCookieSession(cookieName, maxLifeTime, cookieTime)
	p.session = session

}
func (p *Router) Redis(cookieName string, maxLifeTime, cookieTime int64, RedisHost, RedisPwd string) {

	session = NewRedisSession(cookieName, maxLifeTime, cookieTime, RedisHost, RedisPwd)
	p.session = session

}

func (p *Router) Logger(log RouteLogger) {
	logger = log
}

func print(v ...interface{}) {
	if logger != nil {
		logger.Println(v)
	}
}

type HandlerFunc func(*http.Request, *Context)
type HandlerFun func(HandlerFunc) HandlerFunc

func (p *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	method := r.Method

	print("[", method, "]", url)

	if fileLen > 0 && isStatic(url) { //静态资源
		for k, f := range fileRoutes {

			if strings.HasPrefix(url, k) {
				f(w, r)
				return
			}
		}
		print("not found file:", r.URL.String())
		http.Error(w, "Bad file:"+r.URL.String(), http.StatusBadRequest)
	}

	ctx := &Context{Writer: w, Request: r, Sessions: p.session}

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

	nofound := fileRoutes["No-Found-URL-Error-404"]
	if nofound != nil {

		nofound(w, r)
		return
	}

	print("not found URL:", r.URL.String())
	http.Error(w, "Bad URL:"+r.URL.String(), http.StatusBadRequest)

}
func (p *Router) GetSession(r *http.Request, key string) interface{} {

	v, _ := session.Get(r, key)
	return v
}

func isStatic(url string) bool {

	return strings.Contains(url, ".")

}

//执行中间件
func middleware(mws []HandlerFun, h HandlerFunc) HandlerFunc {

	l := len(mws)
	for i := l - 1; i >= 0; i-- {

		h = mws[i](h)
	}
	return h
}

//执行拦截器
func filter(url string, h HandlerFunc) HandlerFunc {

	v := reflect.ValueOf(h)
	fn := runtime.FuncForPC(v.Pointer()).Name()

	for k, f := range mwRoutes {

		if strings.Contains(fn, k) { //按类名
			h = f(h)
		}

		if strings.HasPrefix(url, k) { //按url
			h = f(h)
		}

	}

	return h
}

func (p *Router) add(i int, path string, h HandlerFunc, f ...HandlerFun) {

	url := p.base + path
	m := p.routes[i]

	vh := reflect.ValueOf(h)
	fn := runtime.FuncForPC(vh.Pointer()).Name()

	print(url, " ==> ", fn)

	if len(f) > 0 {

		m[url] = filter(url, middleware(p.mws, f[0](h)))
	} else {

		m[url] = filter(url, middleware(p.mws, h))
	}

	p.rLen[i]++
}

func (p *Router) Static(relativePath string, dirPath string) {

	fileRoutes[relativePath] = func(w http.ResponseWriter, r *http.Request) {

		http.StripPrefix(relativePath, http.FileServer(http.Dir(dirPath))).ServeHTTP(w, r)
	}
	fileLen++
}

func (p *Router) StaticDir(relativePath string, dir string) {

	fileRoutes[relativePath] = func(w http.ResponseWriter, r *http.Request) {

		file := dir + r.URL.Path[1:len(r.URL.Path)]

		info, err := os.Stat(file)
		if err == nil && !info.IsDir() {
			http.ServeFile(w, r, file)
		} else {

			w.WriteHeader(404)
			w.Write([]byte(`not found ` + file))
		}

	}
	fileLen++
}

func (p *Router) File(relativePath string, filePath string, filter ...HandlerFun) {

	var handler = func(req *http.Request, c *Context) {

		info, err := os.Stat(filePath)
		if err == nil && !info.IsDir() {

			http.ServeFile(c.Writer, req, filePath)
		}
	}
	p.add(1, relativePath, handler, filter...)

}

func (p *Router) NoFound(handler http.HandlerFunc) {

	fileRoutes["No-Found-URL-Error-404"] = handler
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

//添加中间件
func (p *Router) Use(h HandlerFun) {

	p.mws = append(p.mws, h)
}

func (p *Router) Any(relativePath string, handler HandlerFunc, filter ...HandlerFun) {

	//p.R(0, relativePath, handler, filter...)
	p.add(1, relativePath, handler, filter...)
	p.add(2, relativePath, handler, filter...)

}

func (p *Router) Get(relativePath string, handler HandlerFunc, filter ...HandlerFun) {

	p.add(1, relativePath, handler, filter...)

}
func (p *Router) Post(relativePath string, handler HandlerFunc, filter ...HandlerFun) {

	p.add(2, relativePath, handler, filter...)

}
func (p *Router) Delete(relativePath string, handler HandlerFunc, filter ...HandlerFun) {

	p.add(3, relativePath, handler, filter...)

}
func (p *Router) Put(relativePath string, handler HandlerFunc, filter ...HandlerFun) {

	p.add(4, relativePath, handler, filter...)

}

func (p *Router) Group(url string, h func(r *Router), params ...HandlerFun) {

	p.base = url
	if len(params) > 0 {
		mwRoutes[url] = params[0]
	}
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

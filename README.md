# gts Web Framework  
golang搭建极简的原生WEB后台项目  

```go

  import (  
    "github.com/gkyh/gts"  
    "net/http"  
   )  
 
   func main() {  
     route := gts.New()  

     route.Cookie("gosessionid", 3600, 0)  

    srv := &http.Server{   
      Addr:           ":8080",   
      Handler:        route,   
      ReadTimeout:    5 * time.Second,   
      WriteTimeout:   10 * time.Second,   
      MaxHeaderBytes: 1 << 20,   
    }   

    //静态文件  
    r.Static("/public", "./public/")     

    r.Use(ws)  
    r.Use(ws2)   
    r.Route("/test", testHandler, HandleIterceptor)  
    r.Group("/group", groupHandler)  
      
    r.Get("/login", func(ctx *gts.Context) {  

      ctx.SetSession("username", "root")
      ctx.WriteString(200, "login")   
    })  

    r.Get("/user", func(c *gts.Context) {  

      user, b := c.SessionVal("username")  
      if b {  
        c.WriteString(200, user.(string))   
      } else {  
        c.WriteString(404, "not found")  
      }  
    })  

    //r.Any("/any", func)  
    //r.Post("/post", func)  
    //r.Get("/get", func)  
    //r.Delete("/delete", func)  
   

    if err := srv.ListenAndServe(); err != nil {

      panic(err)
    }

  }
  func groupHandler(route *gts.Router){  
  
  	route.Get("/test",testFunc)
	route.Post("/test",testFunc)
  }

  func ws(next gts.HandlerFunc) gts.HandlerFunc {  
     return func(ctx *gts.Context) {  

     ip := getRemoteIp(ctx.Request)
      fmt.Println("request ip:" + ip)

	v := map[string]interface{}{
			"reqIP": ip,
	}
	ctx.Set("context", v)
	next(ctx)

    }  
  }  
  

  func HandleIterceptor(next gts.HandlerFunc) gts.HandlerFunc {. 
    return func(c *gts.Context) {   
  
    ip := r.RemoteAddr   
    fmt.Println("handleIterceptor,ip:" + ip)   

    user, _ := c.SessionVal("username")
    if user != nil { 

      fmt.Println(user.(string))   
  
      v := gts.ContextValue{   
        "reqIP":    ip,   
        "username": user.(string),   
      }  
  
      c.Set("context", v)  
      next(c)  
      return  
    }  
  
    c.Redirect("/login")    
    return   
  
  }. 
  }
  ```
  
  ###TestConterller
  
  ```go 
  
  var testHandler *TestConterller = new(TestConterller)   
    
  type TestConterller struct {    
  }   
  //必须在Router方法中添加路由
  func (p *TestConterller) Router(router *gts.Router) {

	  router.Any("/list", p.mlogHandler)  
  
  }  

 func (p *TestConterller) ListHanlder(c *gts.Context) {  
 
 
 
 }  
 ```
 
 Context

```go   
type Context struct {
	Writer   http.ResponseWriter
	Request  *http.Request
	Sessions *Session
}  
```

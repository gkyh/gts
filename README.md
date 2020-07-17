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
    r.Group("/test", testHandler, ws3)

    r.Get("/login", func(w http.ResponseWriter, r *http.Request) {  

      gts.SetSession(w, r, "username", "root")  
      io.WriteString(w, "login")   
    })  

    r.Get("/user", func(w http.ResponseWriter, r *http.Request) {  

      s, b := gts.SessionVal(r, "username")  
      if b {  
        io.WriteString(w, s.(string))   
      } else {  
        io.WriteString(w, "not found")  
      }  
    })  

    //r.Any("/any", func)  
    //r.Post("/any", func)  
    //r.Get("/any", func)  
    //r.Delete("/any", func)  

    if err := srv.ListenAndServe(); err != nil {

      panic(err)
    }

  }


  func ws(next http.HandlerFunc) http.HandlerFunc {  
     return func(w http.ResponseWriter, r *http.Request) {  

      ip := r.RemoteAddr  
      fmt.Println("request ip:" + ip)  

      next(w, r.WithContext(context.WithValue(r.Context(), "reqestIp", ip)))  
      return  

    }  
  }  

  func ws2(next http.HandlerFunc) http.HandlerFunc {   
    return func(w http.ResponseWriter, r *http.Request) {   

      ip := r.RemoteAddr  
      fmt.Println("request ip=========:" + ip)   

      next(w, r.WithContext(context.WithValue(r.Context(), "reqestIp2", ip)))   
      return  

    }  
  }  
  ```
  
  ###TestConterller
  
  ```go 
  
  var testHandler *TestConterller = new(TestConterller)   
    
  type TestConterller struct {    
  }   

  func (p *TestConterller) Router(base string, router *gts.Router, handle ...gts.HandlerFun) {

	  router.Any("/list", p.mlogHandler, handle...)  
  
  }  

 func (p *TestConterller) ListHanlder(w http.ResponseWriter, r *http.Request) {  
 
 
 
 }  
 ```

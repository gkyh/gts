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
    r.Group("/test", testHandler, HandleIterceptor)

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

  func HandleIterceptor(next http.HandlerFunc) http.HandlerFunc {. 
    return func(w http.ResponseWriter, r *http.Request) {   
  
    ip := r.RemoteAddr   
    fmt.Println("handleIterceptor,ip:" + ip)   

    user, _ := gts.SessionVal(r, "username") 
    if user != nil { 

      fmt.Println(user.(string))   
  
      v := gts.ContextValue{   
        "reqIP":    ip,   
        "username": user.(string),   
      }  
  
      ctx := context.WithValue(r.Context(), "context", v)   
      next(w, r.WithContext(ctx))   
      return  
    }  
  
    http.Redirect(w, r, "/login", http.StatusFound)  
    //io.WriteString(w, "on session")  
    return   
  
  }. 
  }
  ```
  
  ###TestConterller
  
  ```go 
  
  var testHandler *TestConterller = new(TestConterller)   
    
  type TestConterller struct {    
  }   

  func (p *TestConterller) Router(base string, router *gts.Router) {

	  router.Any("/list", p.mlogHandler)  
  
  }  

 func (p *TestConterller) ListHanlder(w http.ResponseWriter, r *http.Request) {  
 
 
 
 }  
 ```

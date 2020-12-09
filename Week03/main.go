package main
//未完成
import (
   "context"
   // "errors"
   // "sync"
   // "fmt"
   "net/http"
   "log"
   "io"
   // "os"
   // "os/signal"
   "golang.org/x/sync/errgroup"
)

func main() {
   g, ctx := errgroup.WithContext(context.Background())
   done := make(chan error, 2)
   stop := make(chan struct{})
   g.Go(func () error{
      done <- serveApp(&ctx, stop)
   })
   var stopped bool
   for i:= 0; i < cap(done); i++ {
      if err:=<-done; err != nil {
      fmt.Println("error: %v", err)
      }
      if !stopped {
      stopped = true
      close(stop)
      }
   }
   err := g.Wait()
   if err != nil {
      log.Fatal("out")
   }
}
func serve(addr string, handler http.Handler, stop <- chan struct{}) error {
   s := http.Serve{
      Addr: addr,
      Handler: handler,
   }
   go func() {
      <- stop
      s.Shutdown(context.Background())
   }
   return s.ListenAndServe()
}

func serveApp(w http.ResponseWriter, r *http.Request) {
   fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

func serveAI (w http.ResponseWriter, _ *http.Request) {
   io.WriteString(w, "serveAi")
}

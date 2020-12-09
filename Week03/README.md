# GO 并发

注1:
concurrency 是并发 老师讲反了

老师几年前写了一篇 互斥锁的文章 [文章](https://studygolang.com/articles/1472)


要管好goroutine的生命周期。因为要recover panic（panic 就是挂掉），打印错误。

i++是不是原子的？
内存模型会告诉你原子性，和happen before
如果CPU支持原子自增，多核CPU如何保证原子自增，老式CPU锁总线，所以出来了MESI协议。
Java volatile 的实现
内存屏障是什么？

package sync包的使用
 CAS指令

chan 
go独特的同步机制，通过通讯来实现同步和共享。share memory with community

context
方便做到 级联传递，级联取消

线程和进程的区别

操作系统会为每个应用程序创建一个进程，不同的进程直接相互隔离。

推荐书籍：
<Effective GO>

## Goroutine

### keep yourself busy or do the work yourself

```golang
func main() {
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Hello, GopherCon SG")
  })
  go func() {
    if err:= http.ListenAndServe(":8080", nil); err != nil {
      log.Fatal(err)
    }
  }
  select {}
}
```

缺点：err退出以后main 无感知

```golang
func main() {
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Hello, GopherCon SG")
  })
  if err:= http.ListenAndServe(":8080", nil); err != nil {
    log.Fatal(err)
  }
}
```

缺点： fatal以后 defer无法执行
log.Fatal 一般只在main和init里使用

```golang
func main() {
  mux:= http.newServeMux()
  mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
    fmt.Fprintln(resp, "Hello, QCon")
  })
  go http.ListenAndServe("127.0.0.1:8001", http.DefaultServeMux)
  http.ListenAndServe("0.0.0.0:8080", mux)
}
要知道gor 什么时候退出
要能够主动让 gor 退出

改造：
先抽成两个函数

​```golang
func serveApp() {
  mux:= http.newServeMux()
  mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
    fmt.Fprintln(resp, "Hello, QCon")
  })
  if err := http.ListenAndServe("0.0.0.0:8080", mux); err != nil {
    log.Fatal(err)
  }
}
func serveDebug() {
  uf err := http.ListenAndServe("127.0.0.1:8001", http.DefaultServeMux); err != nil {
    log.Fatal(err)
  }
}
func main() {
  go serveDebug()
  go serveApp()
  select {}
}
```

缺点 还是不能管住gor

```golang
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
func main() {
  done := make(chan error, 2)
  stop := make(chan struct{})
  go func() {
    done <- serveDebug(stop)
  }()
  go func() {
    done <- serveApp(stop)
  }
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
}
```


注意，不要在函数内部偷开gor，比如：

```golang
func main(){
  serve()
}
func serve() {
  go func() {
    //http.xxx
  }()
}
//应该改成
func main(){
  go serve()
}
func serve() {
  //http.xxx
}
```

让调用者来开gor, 除非你已经控制住了你开的这个gor

### Leave concurrency to the caller

func ListDirectory(dir string) ([]string, error)
func ListDirectory(dir string) chan string
第二种写法，无法知道是读完了还是error（可以新建一个struct 解决）
调用者必须持续读取， 无法中断，可能带来性能问题（比如查找目录 提前返回）
filepath.WalkDir也是类似的模型，如果函数启动goroutine ， 必须向调用方提供显示停止该goroutine的方法。通常将异步执行函数的决定权交给该函数的调用方通常更容易
func ListDirectory(root string, workFn func(string)) error

```golang
func leak() {
  ch := make(chan int)
  go func() {
    var := <-ch
    fmt.Println("xxx")
  }()
}
```
gor 泄漏

```golang
func search(term string) (string, error) {
  time.Sleep(200 * time.Millisecond)
  return "some value", nil
}
func process(term string) error {
  record, err:= search()
}
```
超时控制
```golang
func search(term string) (string, error) {
  time.Sleep(200 * time.Millisecond)
  return "some value", nil
}
func process(term string) error {
  ctx, cancel := context.WithTimeout(context.Background(), 100 * time.Millisecond)
  defer cancel()

  ch := make(chan result)

  go func() {
    record
  }()
}
```

```golang
type Tracker struct{}
func (t *Tracker) Event(data string){
  time.Sleep(time.Millisecond)
  log.Println(data)
}
type App struct{
  track Tracker
}
func (a *App) Handle(w ) {
  w.WriteHeader(http.StatusCreated)
  // 无法管控生命周期
  go a.track.Event(" this event")
}
```

改进

```golang
type Tracker struct{}
func (t *Tracker) Event(data string){
  t.wg.Add(1)
  //不建议在request 高频接口里再开gor
  go func() {
    defer t.wg.Done()
    time.Sleep(time.Millisecond)
    log.Println(data)
  }()
}
type App struct{
  track Tracker
}
func (a *App) Handle(w http.ResponseWriter, r *http.Request) {
  w.WriteHeader(http.StatusCreated)
  go a.track.Event(" this event")
}
func () ShutDown(){
  t.wg.wait()
}
func main() {
  var a App
  a.track.Shutdown()
}
```

再改进

```golang
func main() {
  tr := NewTracker()
  go tr.run()
  a.track.Shutdown()
}
type Tracker struct{
  ch chan string
  stop chan struct{}
}
func (t *Tracker) Event(ctx context.Context, data string){
  select {
    case t.ch <- data:
    return nil
    case <-ctx.Done():
    return ctx.Err()
  }
}
func (t *Tracker)Run() {
  for data := range t.ch{
    time.Sleep(1 * time.Millisecond)
    fmt.Println(data)
  }
  t.stop <- struct{}{}
}
func () ShutDown(){
  close(t.ch)
  select {
    case <- t.stop:
    case <- ctx.Done:
  }
}

```

1. 管控生命周期
   1. 知道什么时候结束
   2. 控制他结束
2. 并发扔给调用者

## Memory model

1. https://golang.org/ref/mem 如何保证在一个goroutine 中看到在另一个goroutine修改的变量的值，
 如果程序中修改数据时有其他goroutine同时读取，那么必须将读取串行化。为了串行化访问，请使用channel或其他同步原语，例如sync 和 sync/atomic来保护数据。
 
2. happen before
   
   1. 在**一个**gor中 读和写一定是按照程序中的顺序执行的，即编译器和处理器只有在不会改变这个gor的行为时才可能修改读和写的执行顺序。由于重排，不同的gor可能会看到不同的执行顺序。例如，在一个gor执行a = 1; b = 2; 另一个gor可能看到b在a之前更新
   
3. 重排： 1. 编译器重排 2. CPU重排
   1. 用户写下的代码， 先要编译成汇编代码，为了榨干cpu的性能，比如使用流水线/分支预测等算法，会对读写指令进行重排。这种是内存重排
   
   2. 还有编译器重排，比如
       ```python
       x = 0
       for i in range(100):
        X = 1
     print X
   ```
      
      ```python
      
      x = 1
       for i in range(100):
        print X
      ```
      
      这种情况在多线程下面就会产生bug
   
4. 内核 -》 内存 -》 硬盘 ， L1 L2 L3缓存

   1. 现代CPU为了抚平内核内存硬盘的速度差异，加入了多级缓存，对单线程来说是完美的
   2. 先写入cacheline 不更新内存，多线程情况下会出错

5. 内存屏障 barrier fence 让cache直接扩散，基于此做出各种锁

6. 为了说明读和写的必要条件，我们定义了先行发生（happen before） 如果事件e1 发生在e2之前，我们可以说e2发生在e1之后。如果e1 不发生在e2 之前也不发生在e2之后，我们就说e1 和 e2是并发的。

7. 在单一的独立gor 中先行发生的顺序即是程序中表达的顺序

8. 当下面条件满足时，对变量v的读操作 r 时被允许看到对v的写操作w的：

   1. r不先行发生于w
   2. 在w后r前没有对r的其他写操作

9.  为了保证对变量v的读操作r 看到对v的写操作 w

10. 单个goroutine中没有并发，所以上面两个定义是相同的： 读操作r看到最近一次的写操作w写入v的值

11. 当多个goroutine访问共享变量时，它们必须使用同步事件来建立先行发生这一条件来保证读操作能看到需要的写操作。

12. 对变量v的零值初始化在内存模型中表现的与写操作相同。

13. 对大于single machine word的变量

14. 64位cpu，8字节的 single machine word的变量的读写操作表现的像以不确定顺序
## Package sync
Share Memory By Communicating
go tool compile -S xxx.go 
go build -race

spinning 空转也有优化： intel pause

errgroup
pkg.go.dev/golang.org/x/sync/errgroup

sync.pool
就是用来高频的内存申请，临时对象，适合Request-Driven
作业： 看一下sync.pool 实现： ring buffer + 双向链表
## chan
master-workers
## Package context
其他语言： Thread Local Storage
*a=a{} 重置0值 put
注意defer canel
不要传错ctx
不要带业务逻辑

单飞模式
## References

计算密集型的goroutin 不好打断 退出 但是耗时很短 所以不处理超时
用channel实现一种模式

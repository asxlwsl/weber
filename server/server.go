package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"weber/router"
	"weber/wcontext"
)

// 视图函数signature
// type HandleFunc func(w http.ResponseWriter, r *http.Request)

type Server interface {

	//硬性要求，必须组合http.Handler
	//http.ListenAndServe需要一个Handler对象参数
	http.Handler

	//启动服务
	Start(addr string) error

	//关闭服务
	Stop() error

	//核心
	addRouter(method string, pattern string, handlwFunc wcontext.HandleFunc)
}

type HttpOption func(h *HttpServer)

type HttpServer struct {
	serv *http.Server

	//一个函数类型的属性，用于优雅关闭服务
	stop func() error

	// routers map[string]HandleFunc

	routers *router.Router
}

// 默认的关闭方案
func defaultHttpStop(h *HttpServer) func() error {
	return func() error {
		fmt.Println("== execute default close ==")

		quitSig := make(chan os.Signal)

		//接收到用户终止信号（例如Ctrl+C）,将收到的信号传入quitSig的chan中
		//此处会阻塞，直到接收到信号
		signal.Notify(quitSig, syscall.SIGINT, syscall.SIGTERM)
		<-quitSig

		fmt.Println("== shutdown ==")

		//超时5秒钟的上下文
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)

		defer cancel()

		//关闭之前进行的操作

		//执行http内置的关闭方法，传入ctx上下文
		if err := h.serv.Shutdown(ctx); err != nil {
			log.Fatal("Server shutdown error,", err)
		}

		//关闭之后进行的操作

		select {
		case <-ctx.Done():
			log.Println("timeout of 5 seconds")
		}
		return nil
	}
}

func WithHttpServerStop(fn func() error) HttpOption {
	return func(h *HttpServer) {
		if fn == nil {
			fn = defaultHttpStop(h)
		}
		h.stop = fn
	}
}

// 构造方法
func NewHttpServer(options ...HttpOption) *HttpServer {
	hServer := &HttpServer{
		// routers: map[string]HandleFunc{},
		routers: router.NewRouter(),
	}
	for _, option := range options {
		option(hServer)
	}
	return hServer
}

// 接收客户端请求，转发请求到框架，由框架进行处理
func (h *HttpServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	/*
		//1.路由匹配
		inKey := fmt.Sprintf("%s-%s", request.Method, request.URL)
		if handleFunc := h.routers[inKey]; handleFunc != nil {

			//2.构造当前请求的上下文
			c := wcontext.NewContext(writer, request)
			log.Printf("request %s - %s", c.Method, c.Pattern)

			//3.转发请求
			handleFunc(c)

		} else {
			writer.WriteHeader(http.StatusNotFound)
			writer.Write([]byte("404 not found!"))
		}
	*/
	//生成上下文
	ctx := wcontext.NewContext(writer, request)

	//路由匹配
	handler := h.routers.GetRouter(ctx)

	//调用对应路由处理
	if handler != nil {
		handler(ctx)
	} else if !ctx.Done {
		wcontext.HandleNotFound(ctx)
	}

	ctx.Complete()
}

func (h *HttpServer) Start(addr string) error {
	// return http.ListenAndServe(addr, h)
	httpServer := &http.Server{
		Addr:    addr,
		Handler: h,
	}
	h.serv = httpServer
	return httpServer.ListenAndServe()
}

func (h *HttpServer) Stop() error {
	return h.stop()
}

// 注册路由
// 注册路由的时机
//
//	项目启动的时候进行注册，启动完成后不能再注册
//
// 注册的路由如何存储
//
//	方案一：map[method-pattern]HandleFunc
func (h *HttpServer) addRouter(method string, pattern string, hangleFunc wcontext.HandleFunc) {
	/*
		key := fmt.Sprintf("%s-%s", method, pattern)

		log.Printf("add router %s - %s\n", method, pattern)

		h.routers[key] = hangleFunc
	*/

	h.routers.AddRouter(method, pattern, hangleFunc)
}

/*
func main() {

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		//业务逻辑
	})
	//使用nil会默认使用http.DefaultServerMux
	http.ListenAndServe("8080", nil)
}
*/

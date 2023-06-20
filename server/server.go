package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
	addRouter(method string, pattern string, handlwFunc wcontext.HandleFunc, handleChains ...MiddlewareHandleFunc)

	addGroup(group *RouterGroup)
}

type HttpOption func(h *HttpServer)

type HttpServer struct {
	serv *http.Server

	//一个函数类型的属性，用于优雅关闭服务
	stop func() error

	// routers map[string]HandleFunc

	routers *router.Router

	*RouterGroup

	groups []*RouterGroup
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

	rootGroup := &RouterGroup{}

	var server Server = &HttpServer{
		// routers: map[string]HandleFunc{},
		routers: router.NewRouter(),

		RouterGroup: rootGroup,
	}
	rootGroup.engine = &server
	hServer, ok := server.(*HttpServer)

	if !ok {
		log.Panicln("生成Server失败")
		return nil
	}

	for _, option := range options {
		option(hServer)
	}
	return hServer
}

// 匹配中间件(对应当前URL)
// 中间件在各个路由组上
// 需要在HttpServer上维护整个项目有的路由组
func (h *HttpServer) filterMiddlewares(pattern string) []MiddlewareHandleFunc {

	middlewares := make([]MiddlewareHandleFunc, 0)
	for _, group := range h.groups {
		if strings.HasPrefix(pattern, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	return middlewares
}

func (h *HttpServer) addGroup(group *RouterGroup) {
	h.groups = append(h.groups, group)
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
	// 生成上下文
	ctx := wcontext.NewContext(writer, request)

	// 获取中间价
	middlewares := h.filterMiddlewares(ctx.Pattern)

	// 无论有没有中间件，都将视图函数构建成中间件
	// 将形式统一，以中间件的形式执行
	// 当前请求没有中间件
	if len(middlewares) == 0 {
		middlewares = make([]MiddlewareHandleFunc, 0)
	}

	// 路由匹配
	handler := h.routers.GetRouter(ctx)
	if handler == nil {
		handler = wcontext.HandleNotFound
	}

	// 将匹配到的视图函数添加到mids
	handleFunc := handler

	// 构造责任链(从内向外)
	/*
		- M1
			-M2
				-View
			-M2
		- M1
	*/

	for i := len(middlewares) - 1; i >= 0; i-- {
		handleFunc = middlewares[i](handleFunc)
	}

	handleFunc(ctx)

	/*
		// 调用对应路由处理
		if handler != nil {
			handler(ctx)
		} else if !ctx.Done {
			wcontext.HandleNotFound(ctx)
		}
	*/

	//处理完毕，写入数据
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
func (h *HttpServer) addRouter(method string, pattern string, hangleFunc wcontext.HandleFunc, handleChain ...MiddlewareHandleFunc) {
	/*
		key := fmt.Sprintf("%s-%s", method, pattern)

		log.Printf("add router %s - %s\n", method, pattern)

		h.routers[key] = hangleFunc
	*/

	h.routers.AddRouter(method, pattern, hangleFunc, handleChain...)
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

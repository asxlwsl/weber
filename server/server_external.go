package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/asxlwsl/weber/middleware"
	"github.com/asxlwsl/weber/wcontext"
)

type MiddlewareHandleFunc = middleware.MiddlewareHandleFunc

//路由注册的扩展，提供给用户

type WRoute interface {
	GET(pattern string, handler HandleFunc, handleChains ...MiddlewareHandleFunc)
	POST(pattern string, handler HandleFunc, handleChains ...MiddlewareHandleFunc)
	PUT(pattern string, handler HandleFunc, handleChains ...MiddlewareHandleFunc)
	DELETE(pattern string, handler HandleFunc, handleChains ...MiddlewareHandleFunc)
}

type HandleFunc = wcontext.HandleFunc

// func (h *HttpServer) GET(pattern string, handleFunc HandleFunc) {
// 	h.addRouter(http.MethodGet, pattern, handleFunc)
// }

// func (h *HttpServer) POST(pattern string, handleFunc HandleFunc) {
// 	h.addRouter(http.MethodPost, pattern, handleFunc)
// }

// func (h *HttpServer) DELETE(pattern string, handleFunc HandleFunc) {
// 	h.addRouter(http.MethodDelete, pattern, handleFunc)
// }

//	func (h *HttpServer) PUT(pattern string, handleFunc HandleFunc) {
//		h.addRouter(http.MethodPut, pattern, handleFunc)
//	}

// 统一注册
func (r *RouterGroup) addRouter(method string, pattern string, handler HandleFunc, handleChains ...MiddlewareHandleFunc) {
	pattern = fmt.Sprintf("%s%s", r.prefix, pattern)
	(*r.engine).addRouter(method, pattern, handler, handleChains...)
}

func (r *RouterGroup) GET(pattern string, handleFunc HandleFunc, handleChains ...MiddlewareHandleFunc) {
	r.addRouter(http.MethodGet, pattern, handleFunc, handleChains...)
}

func (r *RouterGroup) POST(pattern string, handleFunc HandleFunc, handleChains ...MiddlewareHandleFunc) {
	r.addRouter(http.MethodPost, pattern, handleFunc, handleChains...)
}

func (r *RouterGroup) DELETE(pattern string, handleFunc HandleFunc, handleChains ...MiddlewareHandleFunc) {
	r.addRouter(http.MethodDelete, pattern, handleFunc, handleChains...)
}

func (r *RouterGroup) PUT(pattern string, handleFunc HandleFunc, handleChains ...MiddlewareHandleFunc) {
	r.addRouter(http.MethodPut, pattern, handleFunc, handleChains...)
}

// 路由组功能
type RouterGroup struct {

	// 路由组前缀（路由组的唯一标识）
	prefix string

	// 路由组的上级路由组（路由组的嵌套）
	parent *RouterGroup

	// 服务实例
	engine *Server

	// 当前路由组的中间件
	middlewares []MiddlewareHandleFunc
}

func (r *RouterGroup) Group(prefix string) *RouterGroup {

	//保险起见，要对prefix进行校验
	/*
		Group("/v1")
		Group("v1/")
		Group("/v1/")
		->
		Group("/v1")
	*/
	prefix = fmt.Sprintf("/%s", strings.Trim(prefix, "/"))

	group := &RouterGroup{prefix: fmt.Sprintf("%s%s", r.prefix, prefix), parent: r, engine: r.engine}

	// 将路由组交给engine维护，用于后续路由组中间价的查找
	(*r.engine).addGroup(group)

	return group
}
func (r *RouterGroup) Run(addr string) {
	(*(r.engine)).Start(addr)
}

// 注册中间件
// 将中间件维护在当前路由组
func (r *RouterGroup) Use(middlewares ...MiddlewareHandleFunc) {
	r.middlewares = append(r.middlewares, middlewares...)
}

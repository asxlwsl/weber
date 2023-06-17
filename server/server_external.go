package server

import (
	"fmt"
	"net/http"
	"strings"
	"weber/wcontext"
)

//路由注册的扩展，提供给用户

type WRoute interface {
	GET(pattern string, handler HandleFunc)
	POST(pattern string, handler HandleFunc)
	PUT(pattern string, handler HandleFunc)
	DELETE(pattern string, handler HandleFunc)
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
func (r *RouterGroup) addRouter(method string, pattern string, handler HandleFunc) {
	pattern = fmt.Sprintf("%s%s", r.prefix, pattern)

	fmt.Println("ADD ROUTER:", pattern)

	(*r.engine).addRouter(method, pattern, handler)
}

func (r *RouterGroup) GET(pattern string, handleFunc HandleFunc) {
	r.addRouter(http.MethodGet, pattern, handleFunc)
}

func (r *RouterGroup) POST(pattern string, handleFunc HandleFunc) {
	r.addRouter(http.MethodPost, pattern, handleFunc)
}

func (r *RouterGroup) DELETE(pattern string, handleFunc HandleFunc) {
	r.addRouter(http.MethodDelete, pattern, handleFunc)
}

func (r *RouterGroup) PUT(pattern string, handleFunc HandleFunc) {
	r.addRouter(http.MethodPut, pattern, handleFunc)
}

// 路由组功能
type RouterGroup struct {

	// 路由组前缀（路由组的唯一标识）
	prefix string

	// 路由组的上级路由组（路由组的嵌套）
	parent *RouterGroup

	// 服务实例
	engine *Server
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

	return &RouterGroup{prefix: fmt.Sprintf("%s%s", r.prefix, prefix), parent: r, engine: r.engine}
}
func (r *RouterGroup) Run(addr string) {
	(*(r.engine)).Start(addr)
}

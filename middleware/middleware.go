package middleware

import "weber/wcontext"

type HandleFunc = wcontext.HandleFunc

// 参数是下一个需要执行的中间件逻辑
// 返回值是当前中间价的逻辑
type MiddlewareHandleFunc func(next HandleFunc) HandleFunc

type HandleChain []MiddlewareHandleFunc

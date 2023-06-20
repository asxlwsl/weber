package middleware

import (
	"fmt"
	"time"
	"weber/wcontext"
)

func Logger() MiddlewareHandleFunc {
	return func(next HandleFunc) HandleFunc {
		return func(ctx *wcontext.Context) {
			ctime := time.Now()
			fmt.Printf("请求进入时间：%v\n", ctime)
			time.Sleep(time.Second * 1)
			next(ctx)
			fmt.Printf("请求返回时间：%v\n", time.Now())
		}
	}
}

func RequestFilter() MiddlewareHandleFunc {
	return func(next HandleFunc) HandleFunc {
		return func(ctx *wcontext.Context) {
			fmt.Println("进入Filter")
			next(ctx)
			fmt.Println("退出Filter")
		}
	}
}

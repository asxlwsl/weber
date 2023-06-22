package middleware

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"
	"weber/wcontext"
)

// 统一将数据写入响应体 需要放在最前
func Flush() MiddlewareHandleFunc {
	return func(next HandleFunc) HandleFunc {
		return func(ctx *wcontext.Context) {
			defer ctx.Complete()
			next(ctx)
		}
	}
}

// 兜底的错误恢复 放在次前
func Recovery() MiddlewareHandleFunc {
	return func(next HandleFunc) HandleFunc {
		return func(ctx *wcontext.Context) {
			defer func() {
				if err := recover(); err != nil {
					message := fmt.Sprintf("%s", err)
					log.Println(Trace(message))
					ctx.Fail(http.StatusInternalServerError, "500 InternalServerError")
				}
			}()
			next(ctx)
		}
	}
}

func Trace(message string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:])

	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}

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

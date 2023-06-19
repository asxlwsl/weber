package tests

import (
	"fmt"
	"testing"
	"time"
	"weber/server"
	"weber/wcontext"
)

func TestHttpServer(t *testing.T) {
	h := server.NewHttpServer(server.WithHttpServerStop(nil))
	go func() {
		err := h.Start(":8080")
		if err != nil {
			fmt.Println("启动失败")
			t.Fail()
		}
		fmt.Println("启动成功")
	}()
	time.Sleep(3 * time.Second)
	err2 := h.Stop()
	if err2 != nil {
		t.Fail()
		fmt.Println("关闭失败")
	}
	fmt.Println("关闭成功")
}
func TestRouterGroup(t *testing.T) {
	h := server.NewHttpServer(server.WithHttpServerStop(nil))
	v1 := h.Group("v1")

	handler := func(ctx *wcontext.Context) {
		ctx.HTML("<h1 style='color:red'>请求成功</h1>")
	}

	v1.Use(server.RequestFilter())
	v1.Use(server.Logger())

	v1.GET("/user", handler)
	v1.POST("/login", handler)

	v1.Run(":8080")
	t.Fail()
	// v2 := v1.Group("v2")

	// v2.GET("/to", handler)

	// v2.Run(":8080")
}

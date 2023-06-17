package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"weber/server"
	"weber/wcontext"
)

func Login(ctx *wcontext.Context) {

	log.Println("Params: ", ctx.Params)

	// ctx.SetResponseBody([]byte("access Login success"))
	ctx.HTML("<h1>Login</h1><h2 style='color:red'>access Login success !</h2>")
	// ctx.JSON(map[string]string{"username": "123", "password": "pwd"})
}
func Register(ctx *wcontext.Context) {
	ctx.SetResponseBody([]byte("access Register success"))
}

func testServer() {
	h := server.NewHttpServer(server.WithHttpServerStop(nil))

	// h.GET("/login/:token", Login)
	h.GET("/login/:username/:password", Login)
	// h.GET("/login/:password/:username", Login)
	h.GET("/login/:username/pop", Login)
	h.GET("/login/:username/push", Login)
	// h.GET("/login/:p/:u/:a", Login)
	h.POST("/register", Register)
	h.GET("/login/toreg/", Register)
	// h.GET("/login/*filepath", Login)
	go func() {
		err := h.Start(":8080")
		if err != nil && http.ErrServerClosed != err {
			fmt.Println("启动失败")
		}
		fmt.Println("启动成功")
	}()
	time.Sleep(3 * time.Second)
	err2 := h.Stop()
	if err2 != nil {

		fmt.Println("关闭失败")
	}
	fmt.Println("关闭成功")
}
func main() {
	// p := router.GetParams("/a/b/c/?", "/a/b/c/?aa=12&cc=vb&pp=4&&k")
	// fmt.Println(p)
	testServer()
}

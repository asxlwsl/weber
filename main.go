package main

import (
	"fmt"
	"net/http"
	"time"
	"weber/server"
)

func Login(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("access Login success"))
}
func Register(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("access Register success"))
}

func main() {
	h := server.NewHttpServer(server.WithHttpServerStop(nil))

	h.GET("/login", Login)
	h.POST("/register", Register)

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

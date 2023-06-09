package tests

import (
	"fmt"
	"testing"
	"time"
	"weber/server"
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

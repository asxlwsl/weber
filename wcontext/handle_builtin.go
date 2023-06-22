package wcontext

import "net/http"

func HandleErrorReturn(errCode int, errMsg string) HandleFunc {
	return func(ctx *Context) {
		ctx.SetStatusCode(errCode)
		ctx.SetResponseBody([]byte(errMsg))
	}
}
func HandleMethodNotAllowed() HandleFunc {
	return HandleErrorReturn(http.StatusMethodNotAllowed, "403 Forbidden!")
}
func HandleNotFound() HandleFunc {
	return HandleErrorReturn(http.StatusNotFound, "404 NOT FOUND!")
}

/*处理静态资源*/
func HandleStaticFile() HandleFunc {

	return func(ctx *Context) {
		
	}
}

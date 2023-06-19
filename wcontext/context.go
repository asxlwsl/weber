package wcontext

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// 请求处理函数
type HandleFunc func(ctx *Context)

// 提供一个新类型，方便操作
type H map[string]any

func HandleErrorReturn(ctx *Context, errCode int, errMsg string) {
	ctx.SetStatusCode(errCode)
	ctx.SetResponseBody([]byte(errMsg))
}

func HandleNotFound(ctx *Context) {
	HandleErrorReturn(ctx, http.StatusNotFound, "404 NOT FOUND!")
}

const (
	CONTENT_TYPE = "Content-Type"
	DEFAULT_CODE = 200
)
const (
	JSON_FORMAT = "application/json;"
	HTML_FORMAT = "text/html;"
	TEXT_FORMAT = "text/plain;"
)

// 对Response进行封装
// ...

// Context 上下文
type Context struct {

	// 响应体对象
	response http.ResponseWriter

	// 请求体对象
	request *http.Request

	// 当前请求的方法
	method string

	// 请求URL
	Pattern string

	//请求相关
	// 1.param参数
	Params map[string]string

	// 2.query参数
	cacheQuery url.Values

	// 3.请求体数据
	cacheBody io.ReadCloser

	// 响应相关

	// 1.状态码
	status int

	// 2.响应头
	header map[string]string

	// 3.响应体
	data []byte

	handlers []HandleFunc

	//当前handler索引
	index int

	//是否继续处理，设为true后不再处理
	Done bool
}

// 获取params参数
func (c *Context) GetQuery(key string) ([]string, error) {
	if c.cacheQuery == nil {
		c.cacheQuery = c.request.URL.Query()
	}
	if value, ok := c.cacheQuery[key]; ok {
		return value, nil
	}
	return nil, QUERY_NOT_FOUND
}

// 获取params参数
func (c *Context) GetParam(key string) (string, error) {

	if value, ok := c.Params[key]; ok {
		return value, nil
	}
	return "", PARAMS_NOT_FOUND
}

// 解析form表单
func (c *Context) GetForm(key string) (string, error) {
	if c.cacheBody == nil {
		c.cacheBody = c.request.Body
	}

	err := c.request.ParseForm()
	if err != nil {
		return "", BODY_NOT_FOUND
	}
	return c.request.PostForm.Get(key), nil
}

// 解析Json格式的请求数据
func (c *Context) BindJSON(dest any) error {
	if c.cacheBody == nil {
		c.cacheBody = c.request.Body
	}
	decoder := json.NewDecoder(c.cacheBody)

	// 不允许未知字段,解析到与dst对应不上的字段会报错
	decoder.DisallowUnknownFields()

	return decoder.Decode(dest)
}

// 获取请求类型
func (c *Context) GetMethod() string {
	return c.method
}

// 设置状态码
func (c *Context) SetStatusCode(code int) {
	c.status = code
}

// 设置响应头
func (c *Context) SetResponseHeader(key string, value string) {
	c.header[key] = value
}

func (c *Context) DelResponseHeader(key string) {
	delete(c.header, key)
}

// 设置响应体
func (c *Context) SetResponseBody(data []byte) {
	c.data = data
}

// 1.响应JSON
func (c *Context) JSON(data any) {
	c.SetStatusCode(DEFAULT_CODE)
	c.SetResponseHeader(CONTENT_TYPE, JSON_FORMAT)
	res, err := json.Marshal(data)
	if err != nil {
		c.SetStatusCode(http.StatusInternalServerError)
		c.DelResponseHeader(CONTENT_TYPE)
		log.Panicln(err)
	}
	c.SetResponseBody(res)
}

// 2.响应HTML
func (c *Context) HTML(html string) {
	c.SetStatusCode(DEFAULT_CODE)
	c.SetResponseHeader(CONTENT_TYPE, HTML_FORMAT)
	c.SetResponseBody([]byte(html))
}

// 3.响应纯文本格式
func (c *Context) TEXT(text string) {
	c.SetStatusCode(DEFAULT_CODE)
	c.SetResponseHeader(CONTENT_TYPE, TEXT_FORMAT)
	c.SetResponseBody([]byte(text))
}

// 处理完成，写入响应数据
func (c *Context) Complete() {

	//写入响应头(先写响应头才会生效)
	for key, value := range c.header {
		c.response.Header().Set(key, value)
	}

	//写入状态码
	if c.status != 0 {
		c.response.WriteHeader(c.status)
	} else {
		c.response.WriteHeader(DEFAULT_CODE)
	}

	//写入响应数据
	c.response.Write(c.data)
}

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	ctx := &Context{
		response: w,
		request:  r,
		method:   r.Method,
		Pattern:  r.URL.Path,
		header:   make(map[string]string),
		handlers: []HandleFunc{},
		index: -1,
		Done:     false,
	}
	parseQueryParams(ctx)
	return ctx
}

// 执行所有视图函数
func (c *Context)Next(){
	c.index++
	size := len(c.handlers)
	for ;c.index<size;c.index++{
		c.handlers[c.index](c)
	}
}



func parseQueryParams(ctx *Context) {
	querys := ctx.request.URL.RawQuery
	params := map[string]string{}
	paramSets := strings.Split(strings.Trim(querys, "?"), "&")

	for _, paramSet := range paramSets {
		idx := strings.IndexRune(paramSet, '=')

		// 防止没有=导致越界
		if idx == -1 {
			if len(paramSet) > 0 {
				params[paramSet] = ""
			}
			continue
		}

		// 防止空值导致越界
		if idx+1 < len(paramSet) {
			params[paramSet[:idx]] = paramSet[idx+1:]
		} else {
			params[paramSet[:idx]] = ""
		}

	}

	ctx.Params = params
}

package goHTTP

import (
	"encoding/json"
	"fmt"
	"net/http"
)

//用于简化json使用
type H map[string]interface{}

//用于描述一次http请求的环境
type Context struct {
	Writer     http.ResponseWriter //向此处写入结果
	Req        *http.Request       //请求体
	Path       string              //请求路径
	Method     string              //请求方法
	Params     map[string]string   //url中动态路由的变量,从路由router中获取
	StatusCode int                 //响应状态码
	handlers   []HandlerFunc       //本次http请求所需要处理的中间件函数的集合
	index      int                 //本次http请求当前所处的中间件函数的位置
}

//新建一个http请求处理的上下文,用于处理http请求
func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer:     w,
		Req:        req,
		Path:       req.URL.Path,
		Method:     req.Method,
		Params:     make(map[string]string),
		StatusCode: 200,
		handlers:   make([]HandlerFunc, 0, 0),
		index:      -1,
	}
}

//执行下一个中间件函数,若中间件函数已全部执行完
func (c *Context) Next() {
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

//跳过所有后续的函数,包括中间件和最终指=执行函数
func (c *Context) Abort() {
	c.index = 65535
}
//获取动态路由中的参数,该参数从router的handle中解析获得
func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}
//从请求体的表单中获得参数
func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}
//从url的query参数中获取,该参数在url尾部
func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}
//设置响应状态码,同时发出响应
func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}
//设置响应头中的参数
func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}
//设置响应码和响应消息,同时将响应内容发出
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}
//以json格式返回响应
func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}
//直接返回数据进行响应
func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}
//构造HTML响应
func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}
//构造失败响应,以json格式进行
func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, H{"message": err})
}

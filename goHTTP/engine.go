package goHTTP

import (
	"log"
	"net/http"
	"strings"
)

//处理函数
//当产生对应的http请求时根据请求类型和url匹配到的处理函数
//调用处理函数去处理http请求
type HandlerFunc func(*Context)

//用于实现ServeHTTP函数,该函数为http的Handler接口下的函数,实现后可自定义http框架
type Engine struct {
	*RouterGroup                //可视为路由群组
	router       *router        //该engine所属路由
	groups       []*RouterGroup // store all groups
}

//新建一个Engine并返回其指针
func New() *Engine {
	engine := &Engine{
		router: newRouter(),
	}
	engine.RouterGroup = &RouterGroup{
		engine: engine,
	}
	engine.groups = []*RouterGroup{
		engine.RouterGroup,
	}
	return engine
}

//添加路由,将对应的请求类型和url与处理函数进行绑定
func (e *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
	//转入调用路由中的添加路由函数
	e.router.addRoute(method, pattern, handler)
}

//GET类型的http请求,用于获取资源
func (e *Engine) GET(pattern string, handler HandlerFunc) {
	e.addRoute("GET", pattern, handler)
}

//POST类型的http请求,用于新建资源
func (e *Engine) POST(pattern string, handler HandlerFunc) {
	e.addRoute("POST", pattern, handler)
}

//DELETE类型的http请求,用于删除资源
func (e *Engine) DELETE(pattern string, handler HandlerFunc) {
	e.addRoute("DELETE", pattern, handler)
}

//PUT类型的http请求,用于更新资源
func (e *Engine) PUT(pattern string, handler HandlerFunc) {
	e.addRoute("PUT", pattern, handler)
}

//http的handler接口的下属函数
//在启动监听时会调用
//主要用于解析请求的路径，查找路由映射表，如果查到，就执行注册的处理方法
func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//中间件的执行函数集合
	var middlewares []HandlerFunc
	//从engine所有下属的group中判断其前缀是否等同于路由组的前缀
	//等同则说明该请求属于该路由组,应当调用该路由组所属的中间件
	for _, group := range e.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	//新建一个请求上下文用于处理该http请求
	c := newContext(w, req)
	//处理请求前先调用其中间件
	c.handlers = middlewares
	//调用路由进行处理
	e.router.handle(c)
}

//调用http库启动监听
func (engine *Engine) Run(addr string) (err error) {
	log.Printf("[hlccd] Listening and serving HTTP on %s", addr)
	return http.ListenAndServe(addr, engine)
}

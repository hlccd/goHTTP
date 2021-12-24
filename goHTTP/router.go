package goHTTP

import (
	"net/http"
	"strings"
)

//路由的结构体
type router struct {
	roots    map[string]*node       //类型结点映射表,建立下属结点与请求类型之间的联系
	handlers map[string]HandlerFunc //路由映射表,建立路由与执行函数之间的联系
}

//新建一个路由并返回其指针
func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

//将url解析为按'/'分层的url局部信息的集合,如果为""则不添加入集合内
//如果某一层的首字符为'*'即动态匹配后续所有则可视为解析完成
func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")
	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' {
				//动态匹配后续所有
				break
			}
		}
	}
	return parts
}

//向该路由下添加一个下属子结点
func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	//解析url
	parts := parsePattern(pattern)
	key := method + "-" + pattern
	_, ok := r.roots[method]
	if !ok {
		//该节点的key不在路由中存在,新建一个进去
		r.roots[method] = &node{}
	}
	//向路由结点下增加结点,同时在路由映射表中添加映射关系
	r.roots[method].insert(pattern, parts, 0)
	r.handlers[key] = handler
}

//从当前路由中获取下属路由结点
func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	//解析url
	searchParts := parsePattern(path)
	root, ok := r.roots[method]
	if !ok {
		//该类请求不存在,直接结束
		return nil, nil
	}
	//从该请求类型中寻找对应的路由结点
	n := root.search(searchParts, 0)
	if n != nil {
		//解析该结点是url前缀
		parts := parsePattern(n.pattern)
		//动态路由参数映射表
		params := make(map[string]string)
		for index, part := range parts {
			if part[0] == ':' {
				//动态匹配,将参数名和参数内容的映射放入映射表内
				params[part[1:]] = searchParts[index]
			}
			if part[0] == '*' && len(part) > 1 {
				//动态路由通配符,将后续所有内容全部添加到映射表内同时结束遍历
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, params
	}
	return nil, nil
}

//该函数调用在engine中,当发生http请求时候会进行调用,同时执行前会先执行中间件
//中间件函数通过context进行了传入
func (r *router) handle(c *Context) {
	//根据请求类型和url信息进行结点获取以及url参数获取
	n, params := r.getRoute(c.Method, c.Path)
	if n != nil {
		//存在该结点,则在中间件的最后添加结点的执行函数
		key := c.Method + "-" + n.pattern
		c.Params = params
		c.handlers = append(c.handlers, r.handlers[key])
	} else {
		//结点不存在,中间件仍然执行
		c.handlers = append(c.handlers, func(c *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		})
	}
	//开始执行中间件
	c.Next()
}

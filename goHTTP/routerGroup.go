package goHTTP

import (
	"log"
	"net/http"
	"path"
)

//路由群结构体
type RouterGroup struct {
	prefix      string        //到当前为止的所有前缀
	middlewares []HandlerFunc //到当前路由群所支持的中间件函数
	parent      *RouterGroup  //该路由群的父路由群
	engine      *Engine       //该路由群所属的engine,该engine可管理其拥有的所有路由群
}

// Group is defined to create a new RouterGroup
// remember all groups share the same Engine instance

//新建一个路由群并返回
//初始时可利用engine创建,即一级路由群
//随后可利用创建的路由群再次进行分组即次级路由群,每一个次级路由群均属于一个高级路由群
//具有从属关系的路由群之间有相同的前缀,即高级路由群前缀是低级路由群的前缀
func (g *RouterGroup) Group(prefix string) *RouterGroup {
	engine := g.engine
	newGroup := &RouterGroup{
		prefix: g.prefix + prefix,
		parent: g,
		engine: engine,
	}
	//创建完成后将新建的路由群加入到engine中
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

//向该路由群添加中间件,同时其所属的子路由群或路由结点也享有该中间件函数
func (g *RouterGroup) Use(middlewares ...HandlerFunc) {
	g.middlewares = append(g.middlewares, middlewares...)
}

//向路由群中添加路由结点
func (g *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := g.prefix + comp
	//通过路由结点进行插入
	g.engine.router.addRoute(method, pattern, handler)
	log.Printf("[hlccd] %-7s - %s", method, pattern)
}

//向当前分组的路由群中增加GET类型的路由结点
func (g *RouterGroup) GET(pattern string, handler HandlerFunc) {
	g.addRoute("GET", pattern, handler)
}

//向当前分组的路由群中增加POST类型的路由结点
func (g *RouterGroup) POST(pattern string, handler HandlerFunc) {
	g.addRoute("POST", pattern, handler)
}

//向当前分组的路由群中增加DELETE类型的路由结点
func (g *RouterGroup) DELETE(pattern string, handler HandlerFunc) {
	g.addRoute("DELETE", pattern, handler)
}

//向当前分组的路由群中增加PUT类型的路由结点
func (g *RouterGroup) PUT(pattern string, handler HandlerFunc) {
	g.addRoute("PUT", pattern, handler)
}

//静态文件,可将目录中的文件映射到路由relativePath
func (g *RouterGroup) Static(relativePath string, root string) {
	handler := g.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	// Register GET handlers
	g.GET(urlPattern, handler)
}

//创建一个静态文件的执行函数
func (g *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(g.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		//检查该文件是否存在
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

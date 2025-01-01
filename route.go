package Nexus

type IRouter interface {
	IRoutes
	Group(string, ...HandlerFunc) *RouterGroup
}

type IRoutes interface {
	Use(...HandlerFunc) IRoutes

	Handle(string, string, ...HandlerFunc) IRoutes
	GET(string, ...HandlerFunc) IRoutes
	POST(string, ...HandlerFunc) IRoutes
	DELETE(string, ...HandlerFunc) IRoutes
	PUT(string, ...HandlerFunc) IRoutes
}

type RouterGroup struct {
	basePath string
	Handlers HandlerFuncList
	engine   *Engine
	root     bool
}

func (r *RouterGroup) ParsePath(method, path string) (handlers HandlerFuncList, params Params, ok bool) {

	// 获取对应方法的路由树根节点
	root := r.engine.trees.get(method)
	if root == nil {
		return nil, nil, false
	}
	var pv nodeValue
	var skipped []skippedNode
	pv = root.getValue(path, &params, &skipped, true)

	if pv.handlers != nil {
		if pv.params != nil {
			return pv.handlers, *pv.params, true
		}
		return pv.handlers, nil, true
	}

	return nil, nil, false
}

var _ IRouter = (*RouterGroup)(nil)

func (r *RouterGroup) Use(middleware ...HandlerFunc) IRoutes {
	r.Handlers = append(r.Handlers, middleware...)
	return r.returnObj()
}
func (r *RouterGroup) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	return &RouterGroup{
		Handlers: r.combineHandlers(handlers),
		basePath: r.calculateAbsolutePath(relativePath),
		engine:   r.engine,
	}
}
func (r *RouterGroup) handle(httpMethod, relativePath string, handlers HandlerFuncList) IRoutes {
	absolutePath := r.calculateAbsolutePath(relativePath)
	handlers = r.combineHandlers(handlers)
	r.engine.addRoute(httpMethod, absolutePath, handlers)
	return r.returnObj()
}

func (r *RouterGroup) Handle(string, string, ...HandlerFunc) IRoutes {
	return r.returnObj()
}

func (r *RouterGroup) GET(path string, handler ...HandlerFunc) IRoutes {
	return r.handle(GET, path, handler)
}

func (r *RouterGroup) POST(path string, handler ...HandlerFunc) IRoutes {
	return r.handle(POST, path, handler)
}

func (r *RouterGroup) PUT(path string, handler ...HandlerFunc) IRoutes {
	return r.handle(PUT, path, handler)
}
func (r *RouterGroup) DELETE(path string, handler ...HandlerFunc) IRoutes {
	return r.handle(DELETE, path, handler)
}

func (r *RouterGroup) returnObj() IRoutes {
	if r.root {
		return r.engine
	}
	return r
}

func (r *RouterGroup) calculateAbsolutePath(relativePath string) string {
	return joinPaths(r.basePath, relativePath)
}

// combineHandlers 合并当前路由组的处理函数和新的处理函数列表。
// 这个方法主要用于内部使用，以合并多个处理函数列表。
// 参数: handlers HandlerFuncList: 待合并的处理函数列表。
// 返回值:  HandlerFuncList: 合并后的处理函数列表。
func (r *RouterGroup) combineHandlers(handlers HandlerFuncList) HandlerFuncList {
	// 计算合并后的处理函数列表的长度。
	finalSize := len(r.Handlers) + len(handlers)
	// 确保处理函数的数量不会超过系统限制。
	assert1(finalSize < int(abortIndex), "too many handlers")

	// 创建一个新的处理函数列表，长度为合并后的长度。
	mergedHandlers := make(HandlerFuncList, finalSize)
	// 将当前路由组的处理函数复制到新的列表中。
	copy(mergedHandlers, r.Handlers)
	// 将新的处理函数列表复制到新的列表中，紧接在当前路由组的处理函数之后。
	copy(mergedHandlers[len(r.Handlers):], handlers)

	// 返回合并后的处理函数列表。
	return mergedHandlers
}

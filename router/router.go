package router

import (
	"fmt"
	"log"
	"strings"
	"weber/middleware"
	"weber/wcontext"

	"golang.org/x/exp/maps"
)

type HandleChain = middleware.HandleChain
type MiddlewareHandleFunc = middleware.MiddlewareHandleFunc

const (
	ROOT_PATH = "/"
)

const (
	DefaultPart = "index"
)

const (
	RNODE  int = 0
	PPNODE int = 1
	PRNODE int = 2
)

func getNodeType(part string) int {
	nType := PPNODE
	switch part[0] {
	case ':':
		nType = PPNODE
	case '*':
		nType = PRNODE
	default:
		nType = RNODE
	}
	log.Println(part, " TYPE:", nType)
	return nType
}

func handleIndexFunc(ctx *wcontext.Context) {
	ctx.SetResponseBody([]byte("服务启动成功，这里是默认页面"))
}

type Router struct {
	roots        map[string]*node
	handlers     map[string]wcontext.HandleFunc
	handleChains map[string]HandleChain
}

// 参数类型的泛型限定 string,map,slice
type ParamType interface {
	string | map[string]string | []string
}

func NewRouter() *Router {
	r := &Router{
		roots:        make(map[string]*node),
		handlers:     make(map[string]wcontext.HandleFunc),
		handleChains: make(map[string]middleware.HandleChain),
	}

	r.roots[ROOT_PATH] = &node{}

	indexNode := &node{pattern: "index", part: "index"}
	r.roots[ROOT_PATH].children = append(r.roots[ROOT_PATH].children, indexNode)
	r.handlers["GET-"+DefaultPart] = handleIndexFunc
	return r
}

func parsePattern(pattern string) []string {
	parts := strings.Split(strings.Trim(pattern, "/"), "/")
	return parts
}

// 将URL的字符串进行切割，分块保存到前缀树上
func (r *Router) AddRouter(method string, pattern string, handler wcontext.HandleFunc, handleChain ...MiddlewareHandleFunc) {

	// _, ok := r.roots[method]

	// //没有对应method的根就创建
	// if !ok {
	// 	r.roots[method] = &node{}
	// }

	parts := parsePattern(pattern)

	//前缀树节点插入
	// r.roots[method].insert(pattern, parts, 0)
	re := r.roots[ROOT_PATH].insert(pattern, parts, 0)

	if !re {
		log.Panicln("{ ", pattern, " } register failed")
	}

	key := fmt.Sprintf("%s-%s", method, pattern)
	r.handlers[key] = handler

	r.handleChains[key] = handleChain

}
func (r *Router) GetRouter(ctx *wcontext.Context) wcontext.HandleFunc {

	n, params := r.getRouter(ctx.GetMethod(), ctx.Pattern)

	maps.Copy(ctx.Params, params)

	// no matched
	if n == nil {
		return nil
	}

	key := fmt.Sprintf("%s-%s", ctx.GetMethod(), n.pattern)

	if fn, ok := r.handlers[key]; ok {

		handleChain, _ := r.handleChains[key]

		for i := len(handleChain) - 1; i >= 0; i-- {
			fn = handleChain[i](fn)
		}
		return fn
	}

	return nil
}

// 获取pattern和query参数
func (r *Router) getRouter(method string, pattern string) (*node, map[string]string) {

	parts := parsePattern(pattern)

	log.Println("pattern: ", pattern, " parts: ", parts)

	// root, ok := r.roots[method]

	// if !ok {
	// 	return nil, nil
	// }

	root, ok := r.roots[ROOT_PATH]

	// 不存在根路径，路由初始化失败
	if !ok {
		return nil, nil
	}

	// 匹配根路由
	if len(parts) == 0 || parts[0] == "" {
		parts = []string{DefaultPart}
	}

	//从顶层开始匹配
	n := root.search(parts, 0)

	//匹配成功，解析参数
	if n != nil {

		return n, GetParams(n.pattern, pattern)
	}

	return nil, nil

}

// 第一个参数为服务段定义的路由，第二个参数为客户端传入
func GetParams(pattern string, URL string) map[string]string {
	params := make(map[string]string)

	formatParts := parsePattern(pattern)

	parseParts := parsePattern(URL)

	rcSize := len(parseParts)

	for idx, part := range formatParts {
		// 匹配 /a/:b模式
		if part[0] == ':' {
			// params[part[1:]] = idx< parseParts[idx]
			if idx < rcSize {
				params[part[1:]] = parseParts[idx]
			} else {
				params[part[1:]] = ""
			}
		} else if part[0] == '*' && len(part) > 1 {
			//匹配带下划线的参数
			// /img/*imgsrc /img/icon/123.jpg -> imgsrc:icon/123.jpg
			if idx < rcSize {
				params[part[1:]] = strings.Join(parseParts[idx:], "/")
			}
		}

	}

	return params
}

// 前缀树/搜索树
type node struct {

	//匹配路由
	pattern string

	//当前节点表示
	part string

	//子节点
	children []*node

	// 参数节点
	paramNode *node

	// 通配节点
	regNode *node

	//匹配模式
	useReg bool
}

//功能

// 1.注册节点:新建node节点,并返回

// 查找第一个匹配节点，辅助插入
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.useReg {
			return child
		}
	}
	return nil
}

// 按层进行递归查找父节点，在父节点后插入
func (n *node) insert(pattern string, parts []string, height int) bool {

	//匹配完成，退出递归
	if len(parts) == height {
		//设置匹配的URL
		n.pattern = pattern

		// log.Println("INSERT COMPLETE: ", n)

		return true
	}
	//获取当前层的part
	part := parts[height]

	nType := getNodeType(part)

	//查找当前层part节点是否存在
	child := n.matchChild(part)

	//存在就处理下一part
	if child != nil {
		return child.insert(pattern, parts, height+1)
	}

	tmpNode := &node{part: part, useReg: part[0] == ':' || part[0] == '*' || part[0] == '?'}

	//判断节点类型
	switch nType {
	case PRNODE:

		// 当前通配符路由，已存在参数路由
		if n.paramNode != nil {
			return false
		}

		//已存在通配符路由
		if n.regNode != nil {
			return false
		}

		//通配符路由设置
		n.regNode = tmpNode

	case PPNODE:
		// 当前参数路由，已存在统配路由
		if n.regNode != nil {
			return false
		}

		// 当前参数节点为空，直接创建该参数节点
		if n.paramNode == nil {
			n.paramNode = tmpNode
		}
		// 参数节点的part与当前part相同
		if n.paramNode.part == part {
			return n.paramNode.insert(pattern, parts, height+1)
		}

		// 存在参数节点，但与当前的参数part不一样，冲突路由，不能再注册
		return false

	case RNODE:
		n.children = append(n.children, tmpNode)
	}
	child = tmpNode
	return child.insert(pattern, parts, height+1)

	/*if child == nil {

		tmpNode := &node{part: part, useReg: part[0] == ':' || part[0] == '*' || part[0] == '?'}
		switch nType {
		case PRNODE:
			if n.regNode != nil {
				return false
			}
			n.regNode = tmpNode
		case PPNODE:
			//已存在参数节点
			if n.paramNode != nil {
				if height+1 < len(parts) {
					tmpNode = n.paramNode
				} else {
					return false
				}
			} else {
				n.paramNode = tmpNode
			}
		case RNODE:
			n.children = append(n.children, tmpNode)
		}
		child = tmpNode
		// child = &node{part: part, useReg: part[0] == ':' || part[0] == '*' || part[0] == '?'}
		// n.children = append(n.children, child)
	}

	return child.insert(pattern, parts, height+1)*/

}

// 2.查找节点匹配
func nodeMatch(n *node, part string) bool {
	return n.part == part
}

// 查找辅助节点
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)

	for _, child := range n.children {
		if nodeMatch(child, part) {
			nodes = append(nodes, child)
		}
	}

	return nodes
}

/*
func (n* node) matchParamNode(part string) []*node{
	// 存储参数路由
	pNodes := make([]*node,0)
	// 存储正则路由
	rNodes := make([]*node,0)

	for _,child := range n.paramNode{
		if child.part[0]==':'{
			pNodes=append(pNodes, child)
		}else if child.part[0]=='*'{
			rNodes =append(rNodes, child)
		}
	}

	//合并路由节点
	for _,child := range rNodes{
		pNodes=append(pNodes, child)
	}
	return pNodes
}
*/

func (n *node) search(parts []string, height int) *node {

	// log.Println("CURR NODE: ", *n, " P: ", parts, " H: ", height)

	if len(parts) == height || strings.HasPrefix(n.part, "*") || strings.HasPrefix(n.part, "?") {

		// 匹配到非叶子节点
		if n.pattern == "" {
			return nil
		}
		return n
	}

	part := parts[height]

	children := n.matchChildren(part)

	for _, child := range children {

		result := child.search(parts, height+1)

		if result != nil {
			return result
		}
	}

	if n.paramNode != nil {
		return n.paramNode.search(parts, height+1)
	}

	return n.regNode
}

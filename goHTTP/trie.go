package goHTTP

//前缀树路由,可匹配":"和"*"两种
//实现了动态路由
import (
	"strings"
)

//前缀树结点
type node struct {
	pattern  string  // 待匹配路由
	part     string  // 路由中当前结点的一部分
	children []*node //	下属子结点,可能出现动态路由的情况
	isWild   bool    // 该节点是否为精确匹配点即动态路由点
}

//根据局部路由信息新建一个结点
//局部url首字符未':'或'*'时可视为动态路由
func newNode(part string) *node {
	return &node{
		pattern:  "",
		part:     part,
		children: make([]*node, 0, 0),
		isWild:   part[0] == ':' || part[0] == '*',
	}
}

//从前缀树结点中插入路由
func (n *node) insert(pattern string, parts []string, idx int) {
	if len(parts) == idx {
		//插入到尾部时设置路由的url
		n.pattern = pattern
		return
	}
	//未插入到重点,找出对应层的string进行继续插入
	part := parts[idx]
	//从子结点中找出匹配该层的结点
	var child *node = nil
	for _, child = range n.children {
		if child.part == part || child.isWild {
			//通过其局部url信息或动态路由找到该结点
			break
		}
	}
	if child == nil {
		//该节点不存在,新建一个结点同时将其添加到当前结点的子结点列表中
		child = newNode(part)
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, idx+1)
}

//从前缀树结点中匹配该url局部信息集合,匹配成功则返回集合终点的结点指针,否则返回nil
//当该局部url的首字符为'*'时可视为匹配成功
//当该结点为动态匹配时可无视其具体string内容直接继续
func (n *node) search(parts []string, height int) *node {
	//根据长度和局部url的首字符进行判断
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			//匹配失败,该结点处无匹配的信息
			return nil
		}
		//匹配成功,返回该结点
		return n
	}
	//从该结点的所有子结点中查找可用于递归查找的结点
	//当局部url信息和当前层string相同时可用于递归查找
	//当该子结点是动态匹配时也可以用于递归查找
	part := parts[height]
	//从所有子结点中找到可用于递归查找的结点
	children := make([]*node, 0, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			//局部string相同或动态匹配
			children = append(children, child)
		}
	}
	for _, child := range children {
		//递归查询,并根据结果进行判断
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}
	return nil
}

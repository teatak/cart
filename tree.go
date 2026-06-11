package cart

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

type nodeType uint8

const (
	static nodeType = iota
	root
	param
	catchAll
)

type node struct {
	path      string
	indices   string
	wildChild bool
	nType     nodeType
	priority  uint32
	children  []*node
	handle    interface{}
}

func (n *node) addRoute(path string, handle interface{}) {
	if path == "" || path[0] != '/' {
		panic("Path must begin with '/' in path '" + path + "'")
	}
	if n.nType == static && n.path == "" && n.handle == nil && len(n.children) == 0 {
		n.nType = root
	}

	segments := splitRoutePath(path)
	cur := n
	for i, segment := range segments {
		kind, name := classifyRouteSegment(segment, path, i == len(segments)-1)
		switch kind {
		case static:
			cur = cur.staticChildOrCreate(segment)
		case param:
			cur = cur.paramChildOrCreate(name, path)
		case catchAll:
			cur = cur.catchAllChildOrCreate(name, path)
		default:
			panic("invalid node type")
		}
	}
	if cur.handle != nil {
		panic("a handle is already registered for path '" + path + "'")
	}
	cur.handle = handle
	n.recomputePriority()
}

func splitRoutePath(path string) []string {
	if path == "/" {
		return nil
	}
	return strings.Split(path[1:], "/")
}

func classifyRouteSegment(segment, fullPath string, last bool) (nodeType, string) {
	if segment == "" {
		return static, segment
	}
	if strings.ContainsAny(segment[1:], ":*") {
		panic("only one wildcard per path segment is allowed, has: '" + segment + "' in path '" + fullPath + "'")
	}
	switch segment[0] {
	case ':':
		name := segment[1:]
		if !validParamName(name) {
			panic("wildcards must be named with a non-empty simple name in path '" + fullPath + "'")
		}
		return param, name
	case '*':
		name := segment[1:]
		if !last {
			panic("catch-all routes are only allowed at the end of the path in path '" + fullPath + "'")
		}
		if !validParamName(name) {
			panic("wildcards must be named with a non-empty simple name in path '" + fullPath + "'")
		}
		return catchAll, name
	default:
		if strings.ContainsAny(segment, ":*") {
			panic("wildcards must occupy a full path segment in path '" + fullPath + "'")
		}
		return static, segment
	}
}

func validParamName(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		if r == '_' || r == '-' || unicode.IsLetter(r) || unicode.IsDigit(r) {
			continue
		}
		return false
	}
	return true
}

func (n *node) staticChild(segment string) *node {
	for _, child := range n.children {
		if child.nType == static && child.path == segment {
			return child
		}
	}
	return nil
}

func (n *node) staticChildOrCreate(segment string) *node {
	if child := n.staticChild(segment); child != nil {
		return child
	}
	child := &node{nType: static, path: segment}
	n.children = append(n.children, child)
	n.rebuildIndices()
	return child
}

func (n *node) paramChild() *node {
	for _, child := range n.children {
		if child.nType == param {
			return child
		}
	}
	return nil
}

func (n *node) paramChildOrCreate(name, fullPath string) *node {
	if child := n.paramChild(); child != nil {
		if child.path != name {
			panic(":" + name + " in new path '" + fullPath + "' conflicts with existing wildcard ':" + child.path + "'")
		}
		return child
	}
	child := &node{nType: param, path: name}
	n.children = append(n.children, child)
	n.rebuildIndices()
	return child
}

func (n *node) catchAllChild() *node {
	for _, child := range n.children {
		if child.nType == catchAll {
			return child
		}
	}
	return nil
}

func (n *node) catchAllChildOrCreate(name, fullPath string) *node {
	if child := n.catchAllChild(); child != nil {
		if child.path != name {
			panic("*" + name + " in new path '" + fullPath + "' conflicts with existing wildcard '*" + child.path + "'")
		}
		return child
	}
	child := &node{nType: catchAll, path: name}
	n.children = append(n.children, child)
	n.rebuildIndices()
	return child
}

func (n *node) rebuildIndices() {
	var b strings.Builder
	for _, child := range n.children {
		if child.nType != static || child.path == "" {
			continue
		}
		b.WriteByte(child.path[0])
	}
	n.indices = b.String()
	n.wildChild = n.paramChild() != nil || n.catchAllChild() != nil
}

func (n *node) recomputePriority() uint32 {
	var priority uint32
	if n.handle != nil {
		priority++
	}
	for _, child := range n.children {
		priority += child.recomputePriority()
	}
	n.priority = priority
	return priority
}

func (n *node) getValue(path string, params func() *Params) (handle interface{}, ps *Params, tsr bool) {
	segments := splitRoutePath(path)
	if match, ok := n.match(segments, 0); ok {
		if params != nil && len(match.params) > 0 {
			ps = params()
			*ps = append(*ps, match.params...)
		}
		return match.handle, ps, false
	}
	if path != "/" {
		alt := path + "/"
		if strings.HasSuffix(path, "/") {
			alt = strings.TrimSuffix(path, "/")
			if alt == "" {
				alt = "/"
			}
		}
		if _, ok := n.match(splitRoutePath(alt), 0); ok {
			return nil, nil, true
		}
	}
	return nil, nil, false
}

type routeMatch struct {
	handle interface{}
	params Params
}

func (n *node) match(segments []string, index int) (routeMatch, bool) {
	if index == len(segments) {
		if n.handle != nil {
			return routeMatch{handle: n.handle}, true
		}
		return routeMatch{}, false
	}

	segment := segments[index]
	if child := n.staticChild(segment); child != nil {
		if match, ok := child.match(segments, index+1); ok {
			return match, true
		}
	}

	if segment != "" {
		if child := n.paramChild(); child != nil {
			if match, ok := child.match(segments, index+1); ok {
				match.params = append(Params{{Key: child.path, Value: segment}}, match.params...)
				return match, true
			}
		}
	}

	if child := n.catchAllChild(); child != nil && child.handle != nil {
		value := "/" + strings.Join(segments[index:], "/")
		return routeMatch{
			handle: child.handle,
			params: Params{{Key: child.path, Value: value}},
		}, true
	}

	return routeMatch{}, false
}

func (n *node) findCaseInsensitivePath(path string, fixTrailingSlash bool) (fixedPath string, found bool) {
	segments := splitRoutePath(path)
	if fixed, ok := n.matchCaseInsensitive(segments, 0); ok {
		if fixed == "" {
			fixed = "/"
		}
		return fixed, true
	}
	if fixTrailingSlash && path != "/" {
		alt := path + "/"
		if strings.HasSuffix(path, "/") {
			alt = strings.TrimSuffix(path, "/")
			if alt == "" {
				alt = "/"
			}
		}
		if fixed, ok := n.matchCaseInsensitive(splitRoutePath(alt), 0); ok {
			if fixed == "" {
				fixed = "/"
			}
			return fixed, true
		}
	}
	return "", false
}

func (n *node) matchCaseInsensitive(segments []string, index int) (string, bool) {
	if index == len(segments) {
		if n.handle != nil {
			return "", true
		}
		return "", false
	}

	segment := segments[index]
	for _, child := range n.children {
		if child.nType != static || !strings.EqualFold(child.path, segment) {
			continue
		}
		if suffix, ok := child.matchCaseInsensitive(segments, index+1); ok {
			return joinFixedPath(child.path, suffix), true
		}
	}

	if segment != "" {
		if child := n.paramChild(); child != nil {
			if suffix, ok := child.matchCaseInsensitive(segments, index+1); ok {
				return joinFixedPath(segment, suffix), true
			}
		}
	}

	if child := n.catchAllChild(); child != nil && child.handle != nil {
		return "/" + strings.Join(segments[index:], "/"), true
	}

	return "", false
}

func joinFixedPath(segment, suffix string) string {
	return "/" + segment + suffix
}

func (n *node) incrementChildPrio(pos int) int {
	return pos
}

func findWildcard(path string) (wilcard string, i int, valid bool) {
	for start, c := range []byte(path) {
		if c != ':' && c != '*' {
			continue
		}
		valid = true
		for end, c := range []byte(path[start+1:]) {
			switch c {
			case '/':
				return path[start : start+1+end], start, valid
			case ':', '*':
				valid = false
			}
		}
		return path[start:], start, valid
	}
	return "", -1, false
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func longestCommonPrefix(a, b string) int {
	i := 0
	max := min(len(a), len(b))
	for i < max && a[i] == b[i] {
		i++
	}
	return i
}

func shiftNRuneBytes(rb [4]byte, n int) [4]byte {
	copy(rb[:], rb[n:])
	return rb
}

func countBytesSkipped(r rune) int {
	return utf8.RuneLen(r)
}

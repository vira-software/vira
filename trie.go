package vira

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

const Version = "v0.1.0"

// Options for the trie
type Options struct {
	// Ignore case when matching URL path
	IgnoreCase bool

	// If enabled, the trie will detect if the current path can't be matched but
	// a handler for the fixed path exists.
	// Matched.FPR will returns either a fixed redirect path or an empty string.
	// For example when "/api/foo" defined and matching "/api//foo",
	// The result Matched.FPR is "/api/foo".
	FixedPathRedirect bool

	// If enabled, the trie will detect if the current path can't be matched but
	// a handler for the path with (without) the trailing slash exists.
	// Matched.TSR will returns either a redirect path or an empty string.
	// For example if /foo/ is requested but a route only exists for /foo, the
	// client is redirected to /foo
	// For example when "/api/foo" defined and matching "/api/foo/",
	// The result Matched.TSR is "/api/foo".
	TrailingSlashRedirect bool
}

// Node represents a node on defined patterns that can be matched.
type Node struct {
	name, allow, pattern, segment, suffix string
	endpoint, wildcard                    bool
	parent                                *Node
	varyChildren                          []*Node
	children                              map[string]*Node
	handlers                              map[string]interface{}
	regex                                 *regexp.Regexp
}

// Trie represents a trie that defining patterns and matching URL.
type Trie struct {
	ignoreCase bool
	fpr        bool
	tsr        bool
	root       *Node
}

// Matched is a result returned by Trie.Match.
type Matched struct {
	// Either a Node pointer when matched or nil
	Node *Node

	// Either a map contained matched values or empty map.
	Params map[string]string

	// If FixedPathRedirect enabled, it may returns a redirect path,
	// otherwise a empty string.
	FPR string

	// If TrailingSlashRedirect enabled, it may returns a redirect path,
	// otherwise a empty string.
	TSR string
}

// the valid characters for the path component:
// [A-Za-z0-9!$%&'()*+,-.:;=@_~]
// http://stackoverflow.com/questions/4669692/valid-characters-for-directory-part-of-a-url-for-short-links
// https://tools.ietf.org/html/rfc3986#section-3.3
var (
	multiSlashReg  = regexp.MustCompile(`/{2,}`)
	wordReg        = regexp.MustCompile(`^\w+$`)
	suffixReg      = regexp.MustCompile(`\+[A-Za-z0-9!$%&'*+,-.:;=@_~]*$`)
	doubleColonReg = regexp.MustCompile(`^::[A-Za-z0-9!$%&'*+,-.:;=@_~]*$`)
	defaultOptions = Options{
		IgnoreCase:            true,
		TrailingSlashRedirect: true,
		FixedPathRedirect:     true,
	}
)

// New return a new Trie
func New(opts ...Options) *Trie {
	options := defaultOptions
	if len(opts) > 0 {
		options = opts[0]
	}

	return &Trie{
		ignoreCase: options.IgnoreCase,
		fpr:        options.FixedPathRedirect,
		tsr:        options.TrailingSlashRedirect,
		root: &Node{
			parent:   nil,
			children: make(map[string]*Node),
			handlers: make(map[string]interface{}),
		},
	}
}

// GetEndpoints returns all endpoints nodes.
func (t *Trie) GetEndpoints() []*Node {
	endpoints := make([]*Node, 0)
	if t.root.endpoint {
		endpoints = append(endpoints, t.root)
	}

	for _, n := range t.root.GetDescendants() {
		if n.endpoint {
			endpoints = append(endpoints, n)
		}
	}

	return endpoints
}

// GetDescendants returns all descendant nodes.
func (n *Node) GetDescendants() []*Node {
	nodes := make([]*Node, 0)
	for _, n := range n.children {
		nodes = append(nodes, n)
		nodes = append(nodes, n.GetDescendants()...)
	}
	for _, n := range n.varyChildren {
		nodes = append(nodes, n)
		nodes = append(nodes, n.GetDescendants()...)
	}
	return nodes
}

// Define define a new pattern on the trie and return the node
//
//	trie := New()
//	node1 := trie.Define("/a")
//	node2 := trie.Define("/a/b")
//	node3 := trie.Define("/a/b")
//	// node2.parent == node1
//	// node2 == node3
//
// The defined pattern can contain three types of parameters:
//
// | Syntax | Description |
// |--------|------|
// | `:name` | named parameter |
// | `:name*` | named with catch-all parameter |
// | `:name(regexp)` | named with regexp parameter |
// | `::name` | not named parameter, it is literal `:name` |
func (t *Trie) Define(pattern string) *Node {
	if strings.Contains(pattern, "//") {
		panic(fmt.Errorf("pattern contains multiple slashes: %s", pattern))
	}

	procPattern := strings.TrimPrefix(pattern, "/")
	if i := strings.IndexRune(procPattern, '?'); i != -1 {
		procPattern = procPattern[:i]
	}

	node := defineNode(t.root, strings.Split(procPattern, "/"), t.ignoreCase)
	if node.pattern == "" {
		node.pattern = pattern
	}

	return node
}

// defineNode defines a new node on the trie
func defineNode(node *Node, patternSeg []string, ignoreCase bool) *Node {
	segment := patternSeg[0]
	segments := patternSeg[1:]
	child := parseNode(node, segment, ignoreCase)

	if len(segments) == 0 {
		child.endpoint = true
		return child
	}

	if child.wildcard {
		panic(fmt.Errorf(`can't define pattern after wildcard: "%s"`, child.getSegments()))
	}

	return defineNode(child, segments, ignoreCase)
}

func parseNode(parent *Node, segment string, ignoreCase bool) *Node {
	originalSegment := segment
	if doubleColonReg.MatchString(segment) {
		originalSegment = segment[1:]
	}
	if ignoreCase {
		originalSegment = strings.ToLower(originalSegment)
	}
	if node := parent.getChild(originalSegment); node != nil {
		return node
	}

	node := &Node{
		segment:  segment,
		parent:   parent,
		children: make(map[string]*Node),
		handlers: make(map[string]interface{}),
	}

	switch {
	case segment == "":
		parent.children[segment] = node
	case doubleColonReg.MatchString(segment):
		// pattern "/a/::" should match "/a/:"
		// pattern "/a/::bc" should match "/a/:bc"
		// pattern "/a/::/bc" should match "/a/:/bc"
		parent.children[originalSegment] = node
	case segment[0] == ':':
		name := segment[1:]

		switch name[len(name)-1] {
		case '*':
			name = name[0 : len(name)-1]
			node.wildcard = true
		default:
			suffix := suffixReg.FindString(name)
			if suffix != "" {
				name = name[0 : len(name)-len(suffix)]
				node.suffix = suffix[1:]
				if node.suffix == "" {
					panic(fmt.Errorf(`invalid pattern: "%s"`, node.getSegments()))
				}
			}

			if name[len(name)-1] == ')' {
				if index := strings.IndexRune(name, '('); index > 0 {
					regex := name[index+1 : len(name)-1]
					if len(regex) > 0 {
						name = name[0:index]
						node.regex = regexp.MustCompile(regex)
					} else {
						panic(fmt.Errorf(`invalid pattern: "%s"`, node.getSegments()))
					}
				}
			}
		}

		// name must be word characters `[0-9A-Za-z_]`
		if !wordReg.MatchString(name) {
			panic(fmt.Errorf(`invalid pattern: "%s"`, node.getSegments()))
		}

		node.name = name
		// check if node exists
		for _, child := range node.varyChildren {
			if child.wildcard {
				if node.wildcard {
					panic(fmt.Errorf(`can't define "%s" after "%s"`, node.getSegments(), child.getSegments()))
				}
				if child.name != node.name {
					panic(fmt.Errorf(`can't define different named parameter "%s" and "%s"`, node.name, child.name))
				}

				return child
			}

			if child.suffix != node.suffix {
				continue
			}

			if node.wildcard && (child.regex == nil && node.regex == nil) || child.regex != nil && node.regex != nil && child.regex.String() == node.regex.String() {
				if child.name != node.name {
					panic(fmt.Errorf(`invalid pattern name "%s", as prev defined "%s"`, node.name, child.getSegments()))
				}

				return child
			}
		}
		parent.varyChildren = append(parent.varyChildren, node)
		if s := parent.varyChildren; len(s) > 1 {
			sort.SliceStable(s, func(i, j int) bool {
				// i > j
				switch {
				case s[i].suffix == "" && s[j].suffix != "":
					return false
				case s[i].suffix != "" && s[j].suffix == "":
					return true
				case s[i].regex != nil && s[j].regex == nil:
					return true
				default:
					return false
				}
			})
		}

	case segment[0] == '*' || segment[0] == '(' || segment[0] == ')':
		panic(fmt.Errorf(`invalid pattern: "%s"`, node.getSegments()))

	case segment[len(segment)-1] == '*':
		node.wildcard = true
		parent.children[originalSegment[0:len(originalSegment)-1]] = node
	default:
		parent.children[originalSegment] = node
	}

	return node
}

func (node *Node) getChild(segment string) *Node {
	if node.children == nil {
		return nil
	}
	return node.children[segment]
}

func (n *Node) getSegments() string {
	segments := n.segment
	if n.parent != nil {
		segments = n.parent.getSegments() + "/" + segments
	}
	return segments
}

// Match try to match path. It will returns a Matched instance
// includes *Node, Params and Tsr when matching success, otherwise nil.
//
//	matched := trie.Match("/a/b")
func (t *Trie) Match(path string) *Matched {
	if path == "" || path[0] != '/' {
		panic(fmt.Errorf(`path is not start with "/": "%s"`, path))
	}
	fixedLen := len(path)
	if t.fpr {
		path = fixPath(path)
		fixedLen -= len(path)
	}

	start := 1
	end := len(path)
	matched := new(Matched)
	parent := t.root

	for i := 1; i <= end; i++ {
		if i < end && path[i] != '/' {
			continue
		}
		segment := path[start:i]
		originalSegment := segment
		if t.ignoreCase {
			originalSegment = strings.ToLower(segment)
		}
		node := matchNode(parent, originalSegment)
		if node == nil {
			// TrailingSlashRedirect: /abc/efg/ -> /abc/efg
			if t.tsr && parent.endpoint && i == end && segment == "" {
				matched.TSR = path[:end-1]
				if t.fpr && fixedLen > 0 {
					matched.FPR = matched.TSR
					matched.TSR = ""
				}
			}
			return matched
		}

		parent = node
		if parent.name != "" {
			if matched.Params == nil {
				matched.Params = make(map[string]string)
			}
			if parent.wildcard {
				matched.Params[parent.name] = path[start:end]
				break
			} else {
				if parent.suffix != "" {
					segment = segment[0 : len(segment)-len(parent.suffix)]
				}
				matched.Params[parent.name] = segment
			}
		}
		start = i + 1
	}

	switch {
	case parent.endpoint:
		matched.Node = parent
		if t.fpr && fixedLen > 0 {
			matched.FPR = path
			matched.Node = nil
		}
	case t.tsr && parent.getChild("") != nil:
		// TrailingSlashRedirect: /abc/efg -> /abc/efg/
		matched.TSR = path + "/"
		if t.fpr && fixedLen > 0 {
			matched.FPR = matched.TSR
			matched.TSR = ""
		}
	}

	return matched
}

func matchNode(parent *Node, segment string) (child *Node) {
	if child = parent.getChild(segment); child != nil || segment == "" {
		return
	}
	for _, child = range parent.varyChildren {
		originalSegment := segment
		if child.suffix != "" {
			if segment == child.suffix || !strings.HasSuffix(segment, child.suffix) {
				continue
			}
			originalSegment = segment[0 : len(segment)-len(child.suffix)]
		}
		if child.regex != nil && !child.regex.MatchString(originalSegment) {
			continue
		}
		return
	}
	return nil
}

func fixPath(path string) string {
	if !strings.Contains(path, "//") {
		return path
	}
	return multiSlashReg.ReplaceAllString(path, "/")
}

// Handle is used to mount a handler with a method name to the node.
//
//	t := New()
//	node := t.Define("/a/b")
//	node.Handle("GET", handler1)
//	node.Handle("POST", handler1)
func (n *Node) Handle(method string, handler interface{}) {
	if n.GetHandler(method) != nil {
		panic(fmt.Errorf(`"%s" already defined`, n.getSegments()))
	}
	n.handlers[method] = handler
	if n.allow == "" {
		n.allow = method
	} else {
		n.allow += ", " + method
	}
}

// GetHandler ...
// GetHandler returns handler by method that defined on the node
//
//	trie := New()
//	trie.Define("/api").Handle("GET", func handler1() {})
//	trie.Define("/api").Handle("PUT", func handler2() {})
//
//	trie.Match("/api").Node.GetHandler("GET").(func()) == handler1
//	trie.Match("/api").Node.GetHandler("PUT").(func()) == handler2
func (n *Node) GetHandler(method string) interface{} {
	return n.handlers[method]
}

// GetAllow returns allow methods defined on the node
//
//	trie := New()
//	trie.Define("/").Handle("GET", handler1)
//	trie.Define("/").Handle("PUT", handler2)
//
//	// trie.Match("/").Node.GetAllow() == "GET, PUT"
func (n *Node) GetAllow() string {
	return n.allow
}

// GetPattern returns pattern defined on the node
func (n *Node) GetPattern() string {
	return n.pattern
}

// GetMethods returns methods defined on the node
func (n *Node) GetMethods() []string {
	methods := make([]string, 0, len(n.handlers))
	for key := range n.handlers {
		methods = append(methods, key)
	}
	return methods
}

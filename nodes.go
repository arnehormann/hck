package hck

import (
	"io"

	"golang.org/x/net/html"
)

type loopError struct{}

func (e loopError) Error() string {
	return "graph contains loops"
}

// index retrieves the index of the first matching node.
// It returns -1 if no match is found.
func index(ns []*Node, m Matcher) int {
	for i := range ns {
		if m.Match(ns[i]) {
			return i
		}
	}
	return -1
}

// splice copies ns, modifies it and retrieves the copy.
// Starting at i, del nodes are replaced with n.
func splice(ns []*Node, i, del int, n ...*Node) []*Node {
	if del < 0 {
		return nil
	}
	dest := make([]*Node, len(ns)+len(n)-del)
	si := i
	di := copy(dest[:si], ns[:si])
	di += del - copy(dest[di:], n)
	di += copy(dest[di:], ns[si:])
	return dest
}

type Siblings []*Node

// Index retrieves the index of the first matching node.
// It returns -1 if no match is found.
func (s Siblings) Index(m Matcher) int {
	return index([]*Node(s), m)
}

// Convert Nodes to /x/net/html.Node siblings.
func (s Siblings) convert(parent *html.Node) (first, last *html.Node) {
	var prev *html.Node
	for _, sib := range s {
		h := sib.convert()
		h.Parent = parent
		h.PrevSibling = prev
		if prev != nil {
			prev.NextSibling = h
		} else {
			first = h
		}
		prev = h
	}
	return first, prev
}

func (s Siblings) Render(w io.Writer) error {
	doc := &html.Node{
		Type: html.DocumentNode,
	}
	first, last := s.convert(doc)
	doc.FirstChild = first
	doc.LastChild = last
	return html.Render(w, doc)
}

// Node is an alternative to golang.org/x/net/html.Node intended for dom mutation.
// It stores a minimal amount of references that have to be updated on transformations.
type Node struct {
	Children  Siblings
	Namespace string
	Data      string
	Attributes
	Type html.NodeType
}

// Convert a /x/net/html.Node to a Node.
func Convert(h *html.Node) *Node {
	var children Siblings
	for c := h.FirstChild; c != nil; c = c.NextSibling {
		children = append(children, Convert(c))
	}
	return &Node{
		Children:   children,
		Namespace:  h.Namespace,
		Data:       h.Data,
		Attributes: Attributes(h.Attr),
		Type:       h.Type,
	}
}

// Clone retrieves a copy of the node.
func (n *Node) Clone() *Node {
	if n == nil {
		return n
	}
	return &Node{
		Children:   append(Siblings{}, n.Children...),
		Namespace:  n.Namespace,
		Data:       n.Data,
		Attributes: append(Attributes{}, n.Attributes...),
		Type:       n.Type,
	}
}

func (n *Node) Swap(n2 *Node) {
	n.Children, n2.Children = n2.Children, n.Children
	n.Namespace, n2.Namespace = n2.Namespace, n.Namespace
	n.Data, n2.Data = n2.Data, n.Data
	n.Attributes, n2.Attributes = n2.Attributes, n.Attributes
	n.Type, n2.Type = n2.Type, n.Type
}

func (n *Node) convert() *html.Node {
	h := &html.Node{
		Namespace: n.Namespace,
		Data:      n.Data,
		Type:      n.Type,
	}
	// normalize strings
	h.DataAtom, h.Data = atomize(n.Data)
	n.Attributes.atomize()
	h.Attr = []html.Attribute(n.Attributes)
	// add children
	h.FirstChild, h.LastChild = n.Children.convert(h)
	return h
}

// Convert a Node to a /x/net/html.Node.
// If a node is an ancestor of its own parent, an error will be returned.
func (n *Node) Convert() (*html.Node, error) {
	if !n.Loopfree() {
		return nil, loopError{}
	}
	return n.convert(), nil
}

func (n *Node) Attribute(key, namespace string) *html.Attribute {
	if n == nil {
		return nil
	}
	return n.Attributes.find(key, namespace)
}

func (n *Node) Attrs() *Attributes {
	return &n.Attributes
}

func (n *Node) Attr(key string) string {
	return n.Attributes.Get(key, "")
}

func (n *Node) AttrNS(key, namespace string) string {
	return n.Attributes.Get(key, namespace)
}

func (n *Node) SetAttr(key, value string) (was string) {
	return n.Attrs().Set(key, "", value)
}

func (n *Node) SetAttrNS(key, namespace, value string) (was string) {
	return n.Attrs().Set(key, namespace, value)
}

func (n *Node) Match(m *Node) bool {
	return n == m
}

func (n *Node) Render(w io.Writer) error {
	doc, err := n.Convert()
	if err != nil {
		return err
	}
	if n.Type != html.DocumentNode {
		tmp := &html.Node{
			Type:       html.DocumentNode,
			FirstChild: doc,
			LastChild:  doc,
		}
		doc.Parent = tmp
		doc = tmp
	}
	return html.Render(w, doc)
}

// Loopfree reports whether no node is the ancestor of its own parents.
func (n *Node) Loopfree() bool {
	return n.loopfree(nil, make(map[*Node]*nodeSeq))
}

func (n *Node) loopfree(parent *Node, parents map[*Node]*nodeSeq) bool {
	if n == nil {
		return true
	}
	known := parents[n]
	if known != nil && known.contains(parent) {
		// n was checked already and exists under a different parent.
		// we have a loop when one parent is reachable from another parent.
		return false
	}
	parents[n] = &nodeSeq{
		next: known,
		node: parent,
	}
	for _, c := range n.Children {
		if !c.loopfree(n, parents) {
			return false
		}
	}
	return true
}

// nodeSeq is a single linked list of nodes
type nodeSeq struct {
	next *nodeSeq
	node *Node
}

func (s *nodeSeq) contains(n *Node) bool {
	if s == nil {
		return false
	}
	if s.node == n {
		return true
	}
	for _, c := range n.Children {
		if s.contains(c) {
			return true
		}
	}
	return s.next.contains(n)
}

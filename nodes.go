package hck

import "golang.org/x/net/html"

type loopError struct{}

func (e loopError) Error() string {
	return "graph contains loops"
}

// Nodes is a sortable node slice
type Nodes []*Node

func (ns Nodes) Len() int {
	return len(ns)
}

func (ns Nodes) Swap(i, j int) {
	ns[i], ns[j] = ns[j], ns[i]
}

func (ns Nodes) Index(m Matcher) int {
	for i := range ns {
		if m.Match(ns[i]) {
			return i
		}
	}
	return -1
}

// Convert Nodes to /x/net/html.Node siblings.
func (ns Nodes) convert(parent *html.Node) (first, last *html.Node) {
	var prev *html.Node
	for _, c := range ns {
		h := c.convert()
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

// Splice retrieves a modified copy.
// Starting at i, del nodes are replaced with n.
func (ns Nodes) Splice(i, del int, n ...*Node) Nodes {
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

func (n *Nodes) Render(w Writer) error {
	doc = &html.Node{
		Type: html.DocumentNode,
	}
	first, last := n.convert(doc)
	doc.FirstChild = first
	doc.LastChild = last
	return html.Render(w, doc)
}

// Node is an alternative to golang.org/x/net/html.Node intended for dom mutation.
// It stores a minimal amount of references that have to be updated on transformations.
type Node struct {
	Children  Nodes
	Namespace string
	Data      string
	*Attributes
	Type html.NodeType
}

// Convert a /x/net/html.Node to a Node.
func Convert(h *html.Node) *Node {
	var children Nodes
	for c := h.FirstChild; c != nil; c = c.NextSibling {
		children = append(children, Convert(c))
	}
	attrs := Attributes(h.Attr)
	return &Node{
		Children:   children,
		Namespace:  h.Namespace,
		Data:       h.Data,
		Attributes: &attrs,
		Type:       h.Type,
	}
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
	h.Attr = []html.Attribute(*n.Attributes)
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
	return n.Attributes.find(key, namespace)
}

func (n *Node) PutAttribute(key, namespace, value string) {
	attr := n.Attributes.findOrAdd(key, namespace)
	attr.Val = value
}

func (n *Node) Match(m *Node) bool {
	return n == m
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

func (n *Node) Render(w Writer) error {
	doc, err := n.Convert()
	if err != nil {
		return err
	}
	if n.Type != html.DocumentNode {
		tmp := &html.Node{
			Type: html.DocumentNode,
			FirstChild: doc,
			LastChild: doc,
		}
		doc.Parent = tmp
		doc = tmp
	}
	return html.Render(w, doc)
}

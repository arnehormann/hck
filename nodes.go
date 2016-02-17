package hck

import (
	"io"

	"golang.org/x/net/html"
)

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

// convert nodes to /x/net/html.Node siblings.
// Nils are skipped.
func (s Siblings) convert(parent *html.Node) (first, last *html.Node) {
	var prev *html.Node
	for _, sib := range s {
		if sib == nil {
			continue
		}
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

// Render nodes to a writer.
// nil nodes are skipped.
func (s Siblings) Render(w io.Writer) error {
	doc := &html.Node{
		Type: html.DocumentNode,
	}
	first, last := s.convert(doc)
	doc.FirstChild = first
	doc.LastChild = last
	return html.Render(w, doc)
}

func (s Siblings) SplitAfter(n *Node) (Siblings, Siblings) {
	i := s.Index(n) + 1
	return s[:i:i], s[i:]
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

// Parse a tree from r.
func Parse(r io.Reader) (*Node, error) {
	dom, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	return Convert(dom), nil
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

// Swap state with another node and retrieve that node.
func (n *Node) Swap(n2 *Node) *Node {
	n.Children, n2.Children = n2.Children, n.Children
	n.Namespace, n2.Namespace = n2.Namespace, n.Namespace
	n.Data, n2.Data = n2.Data, n.Data
	n.Attributes, n2.Attributes = n2.Attributes, n.Attributes
	n.Type, n2.Type = n2.Type, n.Type
	return n2
}

// convert a Node to a /x/net/html.Node.
func (n *Node) convert() *html.Node {
	if n == nil {
		return nil
	}
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
// If a child node is an ancestor of its own parent, an error will be returned.
func (n *Node) Convert() (*html.Node, error) {
	if n.HasCycle() {
		return nil, loopError{}
	}
	return n.convert(), nil
}

// Attribute retrieves a pointer to the Attribute with the given key and namespace.
// If none exists, nil is returned.
func (n *Node) Attribute(key, namespace string) *html.Attribute {
	if n == nil {
		return nil
	}
	return n.Attributes.find(key, namespace)
}

// Attr retrieves the value of an attribute.
func (n *Node) Attr(key string) string {
	return n.Attributes.Get(key, "")
}

// Attr retrieves the value of an attribute with a namespace.
func (n *Node) AttrNS(key, namespace string) string {
	return n.Attributes.Get(key, namespace)
}

func (n *Node) SetAttr(key, value string) (was string) {
	return n.Attributes.Set(key, "", value)
}

func (n *Node) SetAttrNS(key, namespace, value string) (was string) {
	return n.Attributes.Set(key, namespace, value)
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

// HasCycle reports whether any reachable node is the ancestor of its own parents.
func (n *Node) HasCycle() bool {
	return n.hasCycle(nil, make(map[*Node][]*Node))
}

func (n *Node) hasCycle(parent *Node, parents map[*Node][]*Node) bool {
	if n == nil {
		return false
	}
	known := parents[n]
	parents[n] = append(known, parent)
	check := []*Node{n}
	for next := check; len(next) > 0; {
		for _, c := range check {
			if c == nil {
				continue
			}
			for _, k := range known {
				if k == nil {
					continue
				}
				if k == n {
					return true
				}
			}
			next = append(next, c.Children...)
		}
		check = append(check[:0:0], next...)
		next = next[:0]
	}
	return false
}

type loopError struct{}

func (e loopError) Error() string {
	return "graph contains loops"
}

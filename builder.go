package hck

import "golang.org/x/net/html"

type Builder interface {
	// set namespace on current tag
	Namespace(namespace string) Builder

	// add attribute
	Attr(key, value string) Builder

	// add attribute with namespace
	AttrNS(key, namespace, value string) Builder

	// add tag as child, return builder inside child
	Tag(tag string) Builder

	// add textnode as child
	Text(text string) Builder

	// add multiple child nodes
	Children(children ...*Node) Builder

	// leave a child, return builder inside parent
	Leave() Builder

	// retrieve the current node
	Node() *Node

	// retrieve the root node
	Root() *Node
}

type builder struct {
	path Path
}

var _ Builder = builder{}

func Build(n *Node) Builder {
	return builder{path: Path{n}}
}

func Document(children ...*Node) *Node {
	return &Node{
		Type:     html.DocumentNode,
		Children: Siblings(children),
	}
}

func Text(t string) *Node {
	return &Node{
		Type: html.TextNode,
		Data: t,
	}
}

func Tag(tag string) Builder {
	return builder{
		path: Path{
			&Node{
				Type: html.ElementNode,
				Data: tag,
			},
		},
	}
}

func (b builder) Namespace(ns string) Builder {
	b.Node().Namespace = ns
	return b
}

func (b builder) Attr(key, value string) Builder {
	return b.AttrNS(key, "", value)
}

func (b builder) AttrNS(key, namespace, value string) Builder {
	if len(b.path) == 0 {
		panic("illegal builder")
	}
	n := b.Node()
	n.Attributes = append(n.Attributes, html.Attribute{
		Namespace: namespace,
		Key:       key,
		Val:       value,
	})
	return b
}

func (b builder) Text(text string) Builder {
	if len(b.path) == 0 {
		panic("illegal builder")
	}
	n := b.Node()
	n.Children = append(n.Children, Text(text))
	return b
}

func (b builder) Tag(tag string) Builder {
	if len(b.path) == 0 {
		panic("illegal builder")
	}
	c := Tag(tag).Node()
	n := b.Node()
	n.Children = append(n.Children, c)
	return builder{
		path: append(b.path, c),
	}
}

func (b builder) Children(children ...*Node) Builder {
	if len(b.path) == 0 {
		panic("illegal builder")
	}
	n := b.Node()
	n.Children = append(n.Children, children...)
	return b
}

func (b builder) Leave() Builder {
	depth := len(b.path) - 1
	if depth <= 0 {
		panic("no nodes to leave")
	}
	return builder{
		path: b.path[:depth],
	}
}

func (b builder) Node() *Node {
	return b.path.Node()
}

func (b builder) Root() *Node {
	if n := b.path[0]; n != nil {
		return n
	}
	panic("illegal builder")
}

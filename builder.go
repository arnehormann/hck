package hck

import "golang.org/x/net/html"

type Builder interface {
	Namespace(namespace string) Builder
	Attr(key, value string) Builder
	AttrNS(key, namespace, value string) Builder
	AddText(text string) Builder
	AddChildren(children ...*Node) Builder
	AddChild(tag string) Builder
	Leave() Builder
	Build() *Node
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
	b.path.Node().Namespace = ns
	return b
}

func (b builder) node() *Node {
	return b.path.Node()
}

func (b builder) Attr(key, value string) Builder {
	return b.AttrNS(key, "", value)
}

func (b builder) AttrNS(key, namespace, value string) Builder {
	if len(b.path) == 0 {
		panic("illegal builder")
	}
	n := b.path.Node()
	n.Attributes = append(n.Attributes, html.Attribute{
		Namespace: namespace,
		Key:       key,
		Val:       value,
	})
	return b
}

func (b builder) AddText(text string) Builder {
	if len(b.path) == 0 {
		panic("illegal builder")
	}
	n := b.path.Node()
	n.Children = append(n.Children, Text(text))
	return b
}

func (b builder) AddChildren(children ...*Node) Builder {
	if len(b.path) == 0 {
		panic("illegal builder")
	}
	n := b.path.Node()
	n.Children = append(n.Children, children...)
	return b
}

func (b builder) AddChild(tag string) Builder {
	if len(b.path) == 0 {
		panic("illegal builder")
	}
	c := Tag(tag).Build()
	n := b.path.Node()
	n.Children = append(n.Children, c)
	return builder{
		path: append(b.path, c),
	}
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

func (b builder) Build() *Node {
	if n := b.path[0]; n != nil {
		return n
	}
	panic("illegal builder")
}

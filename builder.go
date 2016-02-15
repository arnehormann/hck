package hck

import "golang.org/x/net/html"

type Builder interface {
	Attr(key, value string) Builder
	AttrNS(key, namespace, value string) Builder
	AddText(text string) Builder
	AddChildren(children ...*Node) Builder
	AddChild(tag string) Builder
	AddChildNS(tag, namespace string) Builder
	Leave() Builder
	Build() *Node
}

type builder struct {
	path Path
}

var _ Builder = builder{}

func Tag(tag string) Builder {
	return TagNS(tag, "")
}

func TagNS(tag, namespace string) Builder {
	return builder{
		path: Path{
			&Node{
				Type:      html.ElementNode,
				Data:      tag,
				Namespace: namespace,
			},
		},
	}
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
	n.Children = append(n.Children, &Node{
		Type: html.TextNode,
		Data: text,
	})
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
	return b.AddChildNS(tag, "")
}

func (b builder) AddChildNS(tag, namespace string) Builder {
	if len(b.path) == 0 {
		panic("illegal builder")
	}
	c := &Node{
		Namespace: namespace,
		Data:      tag,
		Type:      html.ElementNode,
	}
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
	n := b.path.Node()
	if n == nil {
		panic("illegal builder")
	}
	return n
}

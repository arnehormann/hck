package hck

import "golang.org/x/net/html"

type Matcher interface {
	Match(*Node) bool
}

type Match func(*Node) bool

func (m Match) Match(n *Node) bool {
	return m(n)
}

type matchNone struct{}

func (_ matchNone) Match(*Node) bool {
	return false
}

func NewMatcher(ms ...Match) Matcher {
	if len(ms) == 0 {
		return matchNone{}
	}
	if len(ms) == 1 {
		return ms[0]
	}
	return matchAll(ms)
}

type matchAll []Match

func (ms matchAll) Match(n *Node) bool {
	for _, m := range ms {
		if !m(n) {
			return false
		}
	}
	return true
}

func MatchAny(ms ...Match) Matcher {
	return matchAny(ms)
}

type matchAny []Match

func (ms matchAny) Match(n *Node) bool {
	for _, m := range ms {
		if m(n) {
			return true
		}
	}
	return false
}

func MatchTag(tag, namespace string) Matcher {
	_, tag = atomize(tag)
	return &matchTag{
		data:      tag,
		namespace: namespace,
	}
}

type matchTag struct {
	data      string
	namespace string
}

func (m *matchTag) Match(n *Node) bool {
	return n.Type == html.ElementNode && n.Namespace == m.namespace && n.Data == m.data
}

func MatchAttribute(key, namespace, value string) Matcher {
	_, key = atomize(key)
	return &matchAttribute{
		Namespace: namespace,
		Key:       key,
		Val:       value,
	}
}

type matchAttribute html.Attribute

func (m *matchAttribute) Match(n *Node) bool {
	attr := n.Attribute(m.Key, m.Namespace)
	return attr != nil && attr.Val == m.Val
}

type matchChild struct {
	m Matcher
}

func MatchChild(m Matcher) Matcher {
	return matchChild{m: m}
}

func (m matchChild) Match(n *Node) bool {
	return n.Children.Index(m.m) >= 0
}

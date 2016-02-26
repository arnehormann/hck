package hck

import "golang.org/x/net/html"

type Matcher interface {
	Match(*Node) bool
}

type Match func(*Node) bool

func (m Match) Match(n *Node) bool {
	return m(n)
}

func Matchers(ms ...Match) []Matcher {
	mm := make([]Matcher, len(ms))
	for i, m := range ms {
		mm[i] = m
	}
	return mm
}

func MatchAll(ms ...Matcher) Matcher {
	return matchAll(ms)
}

type matchAll []Matcher

func (ms matchAll) Match(n *Node) bool {
	for _, m := range ms {
		if !m.Match(n) {
			return false
		}
	}
	return true
}

func MatchAny(ms ...Matcher) Matcher {
	return matchAny(ms)
}

type matchAny []Matcher

func (ms matchAny) Match(n *Node) bool {
	for _, m := range ms {
		if m.Match(n) {
			return true
		}
	}
	return false
}

func MatchTag(tag string) Matcher {
	_, tag = atomize(tag)
	return matchTag(tag)
}

type matchTag string

func (m matchTag) Match(n *Node) bool {
	return n != nil &&
		n.Type == html.ElementNode &&
		n.Data == string(m)
}

func MatchTagNS(tag, namespace string) Matcher {
	_, tag = atomize(tag)
	return &matchTagNS{
		data:      tag,
		namespace: namespace,
	}
}

type matchTagNS struct {
	data      string
	namespace string
}

func (m *matchTagNS) Match(n *Node) bool {
	return n != nil &&
		n.Type == html.ElementNode &&
		n.Namespace == m.namespace &&
		n.Data == m.data
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
	return attr != nil &&
		attr.Val == m.Val
}

type matchChild struct {
	m Matcher
}

func MatchChild(m Matcher) Matcher {
	return matchChild{m: m}
}

func (m matchChild) Match(n *Node) bool {
	return n != nil &&
		n.Children.Index(m.m) >= 0
}

func MatchType(t html.NodeType) Matcher {
	return matchType(t)
}

type matchType html.NodeType

func (m matchType) Match(n *Node) bool {
	return n != nil &&
		n.Type == html.NodeType(m)
}

func MatchNamespace(namespace string) Matcher {
	return matchNamespace{namespace}
}

type matchNamespace struct {
	namespace string
}

func (m matchNamespace) Match(n *Node) bool {
	return n != nil &&
		n.Namespace == m.namespace
}

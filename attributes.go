package hck

import (
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var (
	attrID    = atom.Id.String()
	attrClass = atom.Class.String()
)

func atomize(s string) (atom.Atom, string) {
	// wasteful due to []byte allocs?
	if a := atom.Lookup([]byte(s)); a != 0 {
		return a, a.String()
	}
	return 0, s
}

func (as Attributes) atomize() {
	for i := range as {
		_, as[i].Key = atomize(as[i].Key)
	}
}

type Attributes []html.Attribute

func (as Attributes) Len() int {
	return len(as)
}

func (as Attributes) Swap(i, j int) {
	as[i], as[j] = as[j], as[i]
}

func (as Attributes) Less(i, j int) bool {
	ai, aj := &as[i], &as[j]
	return ai.Namespace < aj.Namespace ||
		ai.Namespace == aj.Namespace &&
			ai.Key < aj.Key ||
		ai.Key == aj.Key &&
			ai.Val < aj.Val
}

func (as Attributes) index(key, namespace string) int {
	for i := range as {
		if as[i].Key == key && as[i].Namespace == namespace {
			return i
		}
	}
	return -1
}

func (as Attributes) find(key, namespace string) *html.Attribute {
	if i := as.index(key, namespace); i >= 0 {
		return &as[i]
	}
	return nil
}

func (as *Attributes) findOrAdd(key, namespace string) *html.Attribute {
	if attr := as.find(key, namespace); attr != nil {
		return attr
	}
	attrs := *as
	attrs = append(attrs, html.Attribute{
		Namespace: namespace,
		Key:       key,
	})
	*as = attrs
	return &attrs[len(attrs)-1]
}

func (as Attributes) Get(key, namespace string) string {
	if i := as.index(key, namespace); i >= 0 {
		return as[i].Val
	}
	return ""
}

func (as *Attributes) Set(key, namespace, value string) (was string) {
	a := as.findOrAdd(key, namespace)
	a.Val, was = value, a.Val
	return was
}

func (as Attributes) ID() string {
	return as.Get(attrID, "")
}

func (as *Attributes) SetID(val string) (was string) {
	return as.Set(attrID, "", val)
}

// Class retrieves the value of the "class" attribute
func (as Attributes) Class() string {
	return as.Get(attrClass, "")
}

func (as *Attributes) SetClass(val string) (was string) {
	return as.Set(attrClass, "", val)
}

// Classes splits the value of the class attribute into the individual classes.
func Classes(val string) []string {
	// WHATWG: a class is a space separated string;
	// split on sequences of space characters as defined in
	// html.spec.whatwg.org/multipage/infrastructure.html#space-character
	var strs []string
	var i, offs int
	for i = 0; i < len(val); i++ {
		switch val[i] {
		case ' ', '\t', '\r', '\n', '\f':
			if offs < i {
				strs = append(strs, val[offs:i])
			}
			offs = i + 1
			continue
		}
	}
	if offs < i {
		strs = append(strs, val[offs:i])
	}
	return strs
}

func (as *Attributes) AddClass(class string) {
	attr := as.findOrAdd(attrClass, "")
	val := " " + class + " "
	for _, c := range Classes(attr.Val) {
		if seek := " " + c + " "; !strings.Contains(val, seek) {
			val += seek[1:]
		}
	}
	attr.Val = "" + val[1:len(val)-1]
}

func (as *Attributes) DelClass(class string) {
	attr := as.find(attrClass, "")
	if attr == nil {
		return
	}
	// TODO " " is not the only separator...
	val := strings.Replace(" "+attr.Val+" ", " "+class, "", -1)
	attr.Val = "" + val[1:]
}

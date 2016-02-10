package hkc

func (n *Node) Find(ms ...Matcher) *Finder {
	var f *Finder
	for i := len(ms) - 1; i >= 0; i-- {
		f = &Finder{
			f: f,
			m: ms[i],
		}
	}
	f.c = n.Cursor()
	return f
}

type Finder struct {
	f *Finder
	c *Cursor
	m Matcher
}

func (f *Finder) Next() *Node {
	if f == nil || f.c == nil {
		return nil
	}
	if n := f.f.Next(); n != nil {
		return nil
	}
	for {
		n := f.c.Next()
		if n == nil {
			f.c = nil
			return nil
		}
		if !f.m.Match(n) {
			continue
		}
		if f.f == nil {
			return n
		}
		f.f.c = n.Cursor()
		n = f.f.Next()
		if n != nil {
			return n
		}
	}
}

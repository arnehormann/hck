package hck

func (n *Node) Find(ms ...Matcher) *Finder {
	if len(ms) == 0 {
		return &Finder{}
	}
	fs := make(finders, len(ms))
	for i := range ms {
		fs[i].Matcher = ms[i]
	}
	fs[0].Cursor = n.Cursor()
	return &Finder{fs}
}

func (r *Node) PathTo(n *Node) Path {
	f := Document(r).Find(n)
	f.next()
	p := f.Path()
	if len(p) == 0 {
		return p
	}
	return p[1:]
}

func (n *Node) Each(m Matcher, f func(n *Node)) {
	fnd := n.Find(m)
	n2 := fnd.Next()
	for n2 != nil {
		n := n2
		n2 = fnd.Next()
		f(n)
	}
}

type finder struct {
	*Cursor
	Matcher
}

type finders []finder

func (fs finders) next() finders {
	if len(fs) == 0 || fs[0].Cursor == nil {
		return fs[:0]
	}
nextInner:
	if fn := fs[1:]; len(fn) > 0 {
		if inner := fn.next(); len(inner) > 0 {
			return fs
		}
	}
	// search match
	f0 := fs[0]
	for {
		n := f0.Next()
		if n == nil {
			f0.Cursor = nil
			return nil
		}
		if !f0.Match(n) {
			continue
		}
		if len(fs) == 1 {
			return fs
		}
		if len(n.Children) > 0 {
			fs[1].Cursor = n.Children[0].Cursor()
			goto nextInner
		}
	}
}

func (fs finders) appendPath(dest Path) Path {
	for _, f := range fs {
		if f.Cursor == nil {
			return dest
		}
		dest = f.appendPath(dest)
	}
	return dest
}

func (fs finders) Node() *Node {
	for i := len(fs) - 1; i >= 0; i-- {
		if c := fs[i].Cursor; c != nil {
			return c.Node()
		}
	}
	return nil
}

func (fs finders) Path() Path {
	return fs.appendPath(nil)
}

type Finder struct {
	finders
}

func (f Finder) Find(m Matcher) *Finder {
	return &Finder{
		finders: append(
			f.finders[:len(f.finders):len(f.finders)],
			finder{
				Matcher: m,
			},
		),
	}
}

func (f Finder) Next() *Node {
	fs := f.next()
	if fs == nil {
		return nil
	}
	f.finders = fs
	return f.Node()
}

// Each iterates over all found nodes and calls do.
func (f *Finder) Each(do func(Path) (stop bool)) {
	var path Path
	var stop bool
	for n := f.Next(); n != nil && !stop; n = f.Next() {
		path = f.appendPath(path[:0])
		stop = do(path)
	}
}

// All retrieves all remaining findable nodes.
func (f *Finder) All() []*Node {
	var ns []*Node
	for n := f.Next(); n != nil; n = f.Next() {
		ns = append(ns, n)
	}
	return ns
}

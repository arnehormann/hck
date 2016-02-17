package hck

func (n *Node) Cursor() *Cursor {
	if n == nil {
		return nil
	}
	return Path{n}.Cursor()
}

// Path contains all nodes traversed to get to its last node.
type Path []*Node

// Index retrieves the index of the first matching node.
// It returns -1 if no match is found.
func (p Path) Index(m Matcher) int {
	return index([]*Node(p), m)
}

// Node retrieves the last node.
// If the path is empty, it returns nil.
func (p Path) Node() *Node {
	for i := len(p) - 1; i >= 0; i-- {
		if n := p[i]; n != nil {
			return n
		}
	}
	return nil
}

func (p Path) Cursor() *Cursor {
	if len(p) == 0 {
		return nil
	}
	idx := -1
	if pi := len(p) - 2; pi >= 0 {
		idx = p[pi].Children.Index(p[pi+1])
	}
	return &Cursor{
		path: p,
		idx:  idx,
	}
}

func (p Path) Prev() Path {
	if len(p) == 0 {
		return p[:0]
	}
	return p[:len(p)-1]
}

// Cursor points to a node and stores the path to it.
// It simplifies navigation on nodes.
//
// You must not modify the path to a cursor and the sibling nodes.
type Cursor struct {
	path Path

	// index of current node in its parent's children
	idx int
}

// Cursor creates a new cursor pointing to the same node.
func (c *Cursor) Cursor() *Cursor {
	return &Cursor{
		path: append(Path{}, c.path...),
		idx:  c.idx,
	}
}

// Depth retrieves the number of ancestor nodes.
func (c *Cursor) Depth() int {
	if d := len(c.path) - 1; d > 0 {
		return d
	}
	return 0
}

func (c *Cursor) appendPath(dest Path) Path {
	return append(dest, c.path...)
}

func (c *Cursor) Path() Path {
	return append(Path{}, c.path...)
}

// Node retrieves the current node.
func (c *Cursor) Node() *Node {
	return c.path[c.Depth()]
}

// PrevSibling moves to and retrieves the previous sibling node.
// If no previous sibling exists, the cursor does not move and nil is returned.
func (c *Cursor) PrevSibling() *Node {
	if c.idx <= 0 || len(c.path) <= 1 {
		return nil
	}
	c.idx--
	pi := len(c.path) - 2
	c.path[pi+1] = c.path[pi].Children[c.idx]
	return c.Node()
}

// NextSibling moves to and retrieves the next sibling node.
// If no next sibling exists, the cursor does not move and nil is returned.
func (c *Cursor) NextSibling() *Node {
	if len(c.path) <= 1 {
		return nil
	}
	pi := len(c.path) - 2
	ns := c.path[pi].Children
	i := c.idx + 1
	if len(ns) <= i {
		return nil
	}
	c.idx = i
	c.path[pi+1] = ns[i]
	return c.Node()
}

// Parent moves to and retrieves the parent node.
// If no parent exists, the cursor does not move and nil is returned.
func (c *Cursor) Parent() *Node {
	if len(c.path) == 1 {
		return nil
	}
	pi := len(c.path) - 2
	p := c.path[pi]
	c.path = c.path[:pi+1]
	if pi > 0 {
		c.idx = c.path[pi-1].Children.Index(p)
	} else {
		c.idx = 0
	}
	return p
}

// Parent moves to and retrieves the first child node.
// If no children exist, the cursor does not move and nil is returned.
func (c *Cursor) FirstChild() *Node {
	n := c.Node()
	if len(n.Children) == 0 {
		return nil
	}
	c.idx = 0
	c.path = append(c.path, n.Children[c.idx])
	return c.Node()
}

// Parent moves to and retrieves the first child node.
// If no children exist, the cursor does not move and nil is returned.
func (c *Cursor) LastChild() *Node {
	n := c.Node()
	if len(n.Children) == 0 {
		return nil
	}
	c.idx = len(n.Children) - 1
	c.path = append(c.path, n.Children[c.idx])
	return c.Node()
}

// Prev moves to and retrieves the depth-first previous node.
// If no previous node exists, the cursor does not move and nil is returned.
func (c *Cursor) Prev() *Node {
	if p := c.PrevSibling(); p != nil {
		return p
	}
	return c.Parent()
}

// Next moves to and retrieves the depth-first next node.
// If no next node exists, the cursor does not move and nil is returned.
func (c *Cursor) Next() *Node {
	if n := c.FirstChild(); n != nil {
		return n
	}
	if n := c.NextSibling(); n != nil {
		return n
	}
	for {
		n := c.Parent()
		if n == nil {
			return nil
		}
		n = c.NextSibling()
		if n != nil {
			return n
		}
	}
}

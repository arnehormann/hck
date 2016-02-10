package hkc

func (n *Node) Cursor() *Cursor {
	return &Cursor{path: []*Node{n}}
}

// Cursor points to a node and stores the path to it.
// It simplifies navigation on nodes.
//
// You must not modify the path to a cursor and the sibling nodes.
type Cursor struct {
	path Nodes
	// index of current node in its parent's children
	idx int
}

// Cursor creates a new cursor pointing to the same node.
func (c *Cursor) Cursor() *Cursor {
	return &Cursor{
		path: append(Nodes{}, c.path...),
		idx:  c.idx,
	}
}

// Depth retrieves the number of ancestor nodes.
func (c *Cursor) Depth() int {
	return len(c.path) - 1
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

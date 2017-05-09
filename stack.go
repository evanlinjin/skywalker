package skywalker

import (
	"errors"
	"github.com/skycoin/cxo/node"
	"github.com/skycoin/cxo/skyobject"
	"github.com/skycoin/skycoin/src/cipher"
	"sync"
)

var gMux sync.Mutex

// Walker represents an object the walks a root's tree.
type Walker struct {
	rpk   cipher.PubKey
	c     *node.Container
	stack []*Obj
}

// NewWalker creates a new walker with given container and root's public key.
func NewWalker(c *node.Container, r cipher.PubKey) (w *Walker, e error) {
	if c == nil {
		e = errors.New("nil container error")
		return
	}
	return
}

type FinderFunc func(v *skyobject.Value) (found bool, e error)

// Advance advances the walker to a child object with a FinderFunc implementation.
// If no objects exists in the internal stack, search root children.
// The input 'fieldStr', if provided, restricts FinderFunc to search in the specified field name of the top-most object,
// to obtain child object.
func (w *Walker) Advance(ff FinderFunc, fieldStr ...string) error {
	gMux.Lock()
	gMux.Unlock()

	if len(w.stack) == 0 {
		// Search from root when nothing is on object stack yet.
		// Obtain root and it's direct children.
		r := w.c.Root(w.rpk)
		if r == nil {
			return ErrRootNotFound
		}
		values, e := r.Values()
		if e != nil {
			return e
		}
		// Loop through direct children of root.
		for _, v := range values {
			// Search for object with provided FinderFunc.
			found, e := ff(v)
			if e != nil {
				return e
			}
			// If object is found, add to stack and return.
			if found {
				obj := NewObj(v)
				w.stack = append(w.stack, obj)
				return nil
			}
		}
	} else {
		// TODO: Complete.
	}
	return ErrObjNotFound
}

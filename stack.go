package skywalker

import (
	"errors"
	"github.com/skycoin/cxo/skyobject"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/cipher/encoder"
	"sync"
)

var gMux sync.Mutex

type Finder func(v *skyobject.Value) (chosen bool)

// Walker represents an object the walks a root's tree.
type Walker struct {
	rpk   cipher.PubKey
	rsk   cipher.SecKey
	c     *skyobject.Container
	stack []*Obj
}

// NewWalker creates a new walker with given container and root's public key.
func NewWalker(c *skyobject.Container, rpk cipher.PubKey, rsk cipher.SecKey) (w *Walker, e error) {
	if c == nil {
		e = errors.New("nil container error")
		return
	}
	w = &Walker{
		rpk: rpk,
		rsk: rsk,
		c:   c,
	}
	return
}

// Size returns the size of the internal stack of walker.
func (w *Walker) Size() int {
	return len(w.stack)
}

// Clear clears the internal stack.
func (w *Walker) Clear() {
	w.stack = []*Obj{}
}

// Helper function. Obtains top-most object from internal stack.
func (w *Walker) top() (*Obj, error) {
	if w.Size() == 0 {
		return nil, ErrEmptyInternalStack
	}
	return w.stack[w.Size()-1], nil
}

// AdvanceFromRoot advances the walker to a child object of the root.
// It uses a Finder implementation to find the child to advance to.
// This function auto-clears the internal stack.
// Input 'p' should be provided with a pointer to the object in which the chosen root's child should deserialize to.
func (w *Walker) AdvanceFromRoot(p interface{}, finder Finder) error {
	gMux.Lock()
	defer gMux.Unlock()

	// Clear the internal stack.
	w.Clear()

	// Search from root when nothing is on object stack yet.
	// Obtain root and it's direct children.
	r := w.c.LastRoot(w.rpk)
	if r == nil {
		return ErrRootNotFound
	}

	// Loop through direct children of root.
	for _, dRef := range r.Refs() {
		// See if it's the object needed with Finder.
		v, e := r.ValueByDynamic(dRef)
		if e != nil {
			return e
		}
		// If object is found, add to stack and return.
		if finder(v) {
			// Deserialize.
			if e := encoder.DeserializeRaw(v.Data(), p); e != nil {
				return e
			}
			obj := NewObj(v, p)
			w.stack = append(w.stack, obj)
			return nil
		}
	}
	return ErrObjNotFound
}

// AdvanceFromRefsField advances from a field of name 'fieldName' and of type 'skyobject.References'.
// It uses a Finder implementation to find the child to advance to.
// Input 'p' should be provided with a pointer to the object in which the chosen child object should deserialize to.
func (w *Walker) AdvanceFromRefsField(fieldName string, p interface{}, finder Finder) error {
	gMux.Lock()
	defer gMux.Unlock()

	// Obtain root.
	r := w.c.LastRoot(w.rpk)
	if r == nil {
		return ErrRootNotFound
	}

	// Obtain top-most object from internal stack.
	obj, e := w.top()
	if e != nil {
		return e
	}

	// Obtain data from top-most object.
	// Obtain field's value and schema name.
	fRefs, fSchemaName, e := obj.GetFieldAsReferences(fieldName)
	if e != nil {
		return e
	}

	// Get Schema of field references.
	schema, e := r.SchemaByName(fSchemaName)
	if e != nil {
		return e
	}

	// Loop through References and apply Finder.
	for i, ref := range fRefs {
		// Create dynamic reference.
		dynamic := skyobject.Dynamic{
			Object: ref,
			Schema: schema.Reference(),
		}
		v, e := r.ValueByDynamic(dynamic)
		if e != nil {
			return e
		}
		// See if it's the object with Finder.
		if finder(v) {
			// Deserialize.
			if e := encoder.DeserializeRaw(v.Data(), p); e != nil {
				return e
			}
			// Add to stack.
			newObj := obj.Generate(v, p, fieldName, i)
			w.stack = append(w.stack, newObj)
			return nil
		}
	}
	return ErrObjNotFound
}

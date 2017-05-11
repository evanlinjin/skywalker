package skywalker

import (
	"errors"
	"fmt"
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
func (w *Walker) peek() (*Obj, error) {
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
	for i, dRef := range r.Refs() {
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
			obj := NewObj(v.Schema().Reference(), p, finder, "", i)
			w.stack = append(w.stack, obj)
			return nil
		}
	}
	return ErrObjNotFound
}

// AdvanceFromRefsField advances from a field of name 'prevFieldName' and of type 'skyobject.References'.
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
	obj, e := w.peek()
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
		// Obtain value from root.
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
			newObj := obj.Generate(v.Schema().Reference(), p, finder, fieldName, i)
			w.stack = append(w.stack, newObj)
			return nil
		}
	}
	return ErrObjNotFound
}

// AdvanceFromRefField advances from a field of name 'prevFieldName' and type 'skyobject.Reference'.
// No Finder is required as field is a single reference.
// Input 'p' should be provided with a pointer to the object in which the chosen child object should deserialize to.
func (w *Walker) AdvanceFromRefField(fieldName string, p interface{}) error {
	gMux.Lock()
	defer gMux.Unlock()

	// Obtain root.
	r := w.c.LastRoot(w.rpk)
	if r == nil {
		return ErrRootNotFound
	}

	// Obtain top-most object from internal stack.
	obj, e := w.peek()
	if e != nil {
		return e
	}

	// Obtain data from top-most object.
	// Obtain field's value and schema name.
	fRef, fSchemaName, e := obj.GetFieldAsReference(fieldName)
	if e != nil {
		return e
	}

	// Get Schema of field reference.
	schema, e := r.SchemaByName(fSchemaName)
	if e != nil {
		return e
	}

	// Create dynamic reference.
	dynamic := skyobject.Dynamic{
		Object: fRef,
		Schema: schema.Reference(),
	}
	// Obtain value from root.
	v, e := r.ValueByDynamic(dynamic)
	if e != nil {
		return e
	}

	// Deserialize.
	if e := encoder.DeserializeRaw(v.Data(), p); e != nil {
		return e
	}
	// Add to internal stack.
	newObj := obj.Generate(v.Schema().Reference(), p, nil, fieldName, -1)
	w.stack = append(w.stack, newObj)
	return nil
}

// AdvanceFromDynamicField advances from a field of name 'prevFieldName' and type 'skyobject.Dynamic'.
// No Finder is required as field is a single reference.
// Input 'p' should be provided with a pointer to the object in which the chosen child object should deserialize to.
func (w *Walker) AdvanceFromDynamicField(fieldName string, p interface{}) error {
	gMux.Lock()
	defer gMux.Unlock()

	// Obtain root.
	r := w.c.LastRoot(w.rpk)
	if r == nil {
		return ErrRootNotFound
	}

	// Obtain top-most object from internal stack.
	obj, e := w.peek()
	if e != nil {
		return e
	}

	// Obtain data from top-most object.
	// Obtain field's value and schema name.
	fDyn, e := obj.GetFieldAsDynamic(fieldName)
	if e != nil {
		return e
	}

	// Obtain value from root.
	v, e := r.ValueByDynamic(fDyn)
	if e != nil {
		return e
	}

	// Deserialize.
	if e := encoder.DeserializeRaw(v.Data(), p); e != nil {
		return e
	}
	// Add to internal stack.
	newObj := obj.Generate(v.Schema().Reference(), p, nil, fieldName, -1)
	w.stack = append(w.stack, newObj)
	return nil
}

// Retreat retreats one from the internal stack.
func (w *Walker) Retreat() {
	switch w.Size() {
	case 0:
		return
	case 1:
		w.stack = []*Obj{}
	default:
		w.stack = w.stack[:len(w.stack)-1]
		w.stack[len(w.stack)-1].next = nil
	}
}

// TODO: Implement.
// AppendToRefsField
// ReplaceInRefsField
// DeleteInRefsField
//
// ReplaceInRefField replaces the reference field of the top-most object with a new reference; one that is automatically
// generated when saving the object 'p' points to, in the container. This recursively replaces all the associated
// "references" of the object tree and hence, changes the root.
func (w *Walker) ReplaceInRefField(fieldName string, p interface{}) error {
	gMux.Lock()
	defer gMux.Unlock()

	// Obtain top-most object.
	tObj, e := w.peek()
	if e != nil {
		return e
	}

	// Save new obj.
	nRef := w.c.Save(p)
	if e := tObj.ReplaceReferenceField(fieldName, nRef); e != nil {
		return e
	}

	// Recursively save.
	lRef, e := tObj.Save(w.c)
	if e != nil {
		return e
	}

	// Replace root's direct child.
	r := w.c.LastRoot(w.rpk)
	rDyns := r.Refs()
	rDyns[w.stack[0].prevInFieldIndex] = lRef
	r.Replace(rDyns)
	return nil
}

// ReplaceInDynamicField functions the same as 'ReplaceInRefField'. However, it replaces a dynamic reference field other
// than a static reference field.
func (w *Walker) ReplaceInDynamicField(fieldName string, p interface{}) error {
	gMux.Lock()
	defer gMux.Unlock()

	// Obtain top-most object.
	tObj, e := w.peek()
	if e != nil {
		return e
	}

	// Save new object.
	nDyn := w.c.Dynamic(p)
	if e := tObj.ReplaceDynamicField(fieldName, nDyn); e != nil {
		return e
	}

	// Recursively save.
	lDyn, e := tObj.Save(w.c)
	if e != nil {
		return e
	}

	// Replace root's direct child.
	r := w.c.LastRoot(w.rpk)
	rDyns := r.Refs()
	rDyns[w.stack[0].prevInFieldIndex] = lDyn
	r.Replace(rDyns)
	return nil
}

func (w *Walker) String() (out string) {
	tabs := func(n int) {
		for i := 0; i < n; i++ {
			out += "\t"
		}
	}
	out += fmt.Sprint("Root")
	size := w.Size()
	if size == 0 {
		return
	}
	out += fmt.Sprintf(".Refs[%d] ->\n", w.stack[0].prevInFieldIndex)
	for i, obj := range w.stack {
		schName := ""
		s, _ := w.c.CoreRegistry().SchemaByReference(obj.s)
		if s != nil {
			schName = s.Name()
		}

		tabs(i)
		out += fmt.Sprintf("  %s", schName)
		out += fmt.Sprintf(` = "%v"`+"\n", obj.p)

		tabs(i)
		if obj.next != nil {
			out += fmt.Sprintf("  %s", schName)
			out += fmt.Sprintf(".%s", obj.next.prevFieldName)
			if obj.next.prevInFieldIndex != -1 {
				out += fmt.Sprintf("[%d]", obj.next.prevInFieldIndex)
			}
			out += fmt.Sprint(" ->\n")
		}
	}
	return
}

package skywalker

import (
	"errors"
	"github.com/skycoin/cxo/node"
	"github.com/skycoin/cxo/skyobject"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/cipher/encoder"
	"sync"
	//"reflect"
	"fmt"
)

var gMux sync.Mutex

type FinderFunc func(v *skyobject.Value) (found bool, e error)

// Walker represents an object the walks a root's tree.
type Walker struct {
	rpk   cipher.PubKey
	rsk   cipher.SecKey
	c     *node.Container
	stack []*Obj
}

// NewWalker creates a new walker with given container and root's public key.
func NewWalker(c *node.Container, rpk cipher.PubKey, rsk cipher.SecKey) (w *Walker, e error) {
	if c == nil {
		e = errors.New("nil container error")
		return
	}
	w = &Walker{
		rpk: rpk,
		rsk: rsk,
		c: c,
	}
	return
}

func (w *Walker) Size() int {
	return len(w.stack)
}

// Advance advances the walker to a child object with a FinderFunc implementation.
// If no objects exists in the internal stack, search root children.
// The input 'fieldStr', if provided, restricts FinderFunc to search in the specified field name of the top-most object,
// to obtain child object.
func (w *Walker) Advance(ff FinderFunc, p interface{}, fieldStr ...string) error {
	gMux.Lock()
	defer gMux.Unlock()

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
			// See if it's the object needed with FinderFunc.
			found, e := ff(v)
			if e != nil {
				return e
			}
			// If object is found, add to stack and return.
			if found {
				// Deserialize.
				if e := encoder.DeserializeRaw(v.Data(), p); e != nil {
					return e
				}
				obj := NewObj(v, p)
				w.stack = append(w.stack, obj)
				return nil
			}
		}
	} else {
		// TODO: Complete.
		// If fieldStr is not defined, return an error.
		if len(fieldStr) < 1 {
			return ErrFieldNotProvided
		}
		// Obtain top-most object.
		obj := w.stack[w.Size()-1]

		// Get SchemaName of field's array values.
		fVal, e := obj.v.FieldByName(fieldStr[0])
		if e != nil {
			return e
		}
		fmt.Println(fVal.Schema().Name())

		// Get field of obj.
		//objP := reflect.ValueOf(obj.p).Elem()
		//typP := objP.Type()
		//for i := 0; i < objP.NumField(); i++ {
		//	f := objP.Field(i)
		//	if fieldStr[0] == typP.Field(i).Name {
		//		refs := f.Interface().(skyobject.References)
		//		for j, ref := range refs {
		//			v, e := w.c.GetObject("", ref)
		//			if e != nil {
		//				return e
		//			}
		//
		//
		//		}
		//
		//	}
		//}

	}
	return ErrObjNotFound
}

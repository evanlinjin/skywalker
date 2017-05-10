package skywalker

import (
	"github.com/skycoin/cxo/skyobject"
	"reflect"
	"strings"
)

type Obj struct {
	prev *Obj
	next *Obj

	value *skyobject.Value
	p     interface{}

	prevFinder       Finder // Finder used on prev obj used to find current.
	prevFieldName    string // Field name of prev obj used to find current.
	prevInFieldIndex int    // Index of prev obj's field's prevInFieldIndex. -1 if single reference (not array).
}

func NewObj(v *skyobject.Value, p interface{}, finder Finder, fn string, i int) *Obj {
	return &Obj{
		value:            v,
		p:                p,
		prevFinder:       finder,
		prevFieldName:    fn,
		prevInFieldIndex: i,
	}
}

func (o *Obj) Generate(v *skyobject.Value, p interface{}, finder Finder, fn string, i int) *Obj {
	newO := NewObj(v, p, finder, fn, i)
	newO.prev = o
	o.next = newO
	return newO
}

func (o *Obj) Elem() reflect.Value {
	return reflect.ValueOf(o.p).Elem()
}

func (o *Obj) GetFieldAsReferences(fieldName string) (
	refs skyobject.References, schemaName string, e error,
) {
	v := o.Elem()
	vt := v.Type()

	// Obtain field.
	ft, has := vt.FieldByName(fieldName)
	if has == false {
		e = ErrFieldNotFound
		return
	}
	// Check type of field.
	if ft.Type.Kind().String() != "slice" {
		e = ErrFieldHasWrongType
		return
	}

	// Obtain schemaName from field tag.
	fStr := ft.Tag.Get("skyobject")
	schemaName = strings.TrimPrefix(fStr, "schema=")

	// Obtain field value.
	f := v.FieldByName(fieldName)
	refs = f.Interface().(skyobject.References)
	return
}

func (o *Obj) GetFieldAsReference(fieldName string) (
	ref skyobject.Reference, schemaName string, e error,
) {
	v := o.Elem()
	vt := v.Type()

	// Obtain field.
	ft, has := vt.FieldByName(fieldName)
	if has == false {
		e = ErrFieldNotFound
		return
	}
	// Check type of field.
	if ft.Type.Kind().String() != "array" {
		e = ErrFieldHasWrongType
		return
	}

	// Obtain schemaName from field tag.
	fStr := ft.Tag.Get("skyobject")
	schemaName = strings.TrimPrefix(fStr, "schema=")

	// Obtain field value.
	f := v.FieldByName(fieldName)
	ref = f.Interface().(skyobject.Reference)
	return
}

func (o *Obj) GetFieldAsDynamic(fieldName string) (
	dyn skyobject.Dynamic, e error,
) {
	v := o.Elem()
	vt := v.Type()

	// Obtain field.
	ft, has := vt.FieldByName(fieldName)
	if has == false {
		e = ErrFieldNotFound
		return
	}
	// Check type of field.
	if ft.Type.Kind().String() != "struct" {
		e = ErrFieldHasWrongType
		return
	}

	// Obtain field value.
	f := v.FieldByName(fieldName)
	dyn = f.Interface().(skyobject.Dynamic)
	return
}

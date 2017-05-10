package skywalker

import (
	"github.com/skycoin/cxo/skyobject"
	"reflect"
	"strings"
	"fmt"
)

type Obj struct {
	prev *Obj
	next *Obj

	s skyobject.SchemaReference
	p interface{}

	prevFinder       Finder // Finder used on prev obj used to find current.
	prevFieldName    string // Field name of prev obj used to find current.
	prevInFieldIndex int    // Index of prev obj's field's prevInFieldIndex. -1 if single reference (not array).
}

func NewObj(s skyobject.SchemaReference, p interface{}, finder Finder, fn string, i int) *Obj {
	return &Obj{
		s:                s,
		p:                p,
		prevFinder:       finder,
		prevFieldName:    fn,
		prevInFieldIndex: i,
	}
}

func (o *Obj) Generate(s skyobject.SchemaReference, p interface{}, finder Finder, fn string, i int) *Obj {
	newO := NewObj(s, p, finder, fn, i)
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

func (o *Obj) GetSchema(ct *skyobject.Container) skyobject.Schema {
	s, _ := ct.CoreRegistry().SchemaByReference(o.s)
	return s
}

func (o *Obj) ReplaceReferencesField(fieldName string, newRefs skyobject.References) (e error) {
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

	v.FieldByName(fieldName).Set(reflect.ValueOf(newRefs))
	return
}

func (o *Obj) ReplaceReferenceField(fieldName string, newRef skyobject.Reference) (e error) {
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

	v.FieldByName(fieldName).Set(reflect.ValueOf(newRef))
	return
}

func (o *Obj) ReplaceDynamicField(fieldName string, newDyn skyobject.Dynamic) (e error) {
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

	v.FieldByName(fieldName).Set(reflect.ValueOf(newDyn))
	return
}

func (o *Obj) Save(ct *skyobject.Container) (skyobject.Reference, error) {
	fmt.Println(o.p)
	ref := ct.Save(o.p)

	if o.prev == nil {
		return ref, nil
	}

	// Get previous object's field type.
	v := o.prev.Elem()
	vt := v.Type()

	sf, has := vt.FieldByName(o.prevFieldName)
	if has == false {
		return ref, ErrFieldNotFound
	}

	switch sf.Type.Kind().String() {
	case "slice": // skyobject.References
		tRefs, _, e := o.prev.GetFieldAsReferences(o.prevFieldName)
		if e != nil {
			return ref, e
		}
		tRefs[o.prevInFieldIndex] = ref
		e = o.prev.ReplaceReferencesField(o.prevFieldName, tRefs)
		if e != nil {
			return ref, e
		}
	case "array": // skyobject.Reference
		tRef, _, e := o.prev.GetFieldAsReference(o.prevFieldName)
		if e != nil {
			return ref, e
		}
		tRef = ref
		e = o.prev.ReplaceReferenceField(o.prevFieldName, tRef)
		if e != nil {
			return ref, e
		}
	case "struct": // skyobject.Dynamic
		tDyn, e := o.prev.GetFieldAsDynamic(o.prevFieldName)
		if e != nil {
			return ref, e
		}
		tDyn.Object = ref
		e = o.prev.ReplaceDynamicField(o.prevFieldName, tDyn)
		if e != nil {
			return ref, e
		}
	}

	return o.prev.Save(ct)
}

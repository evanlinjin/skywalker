package skywalker

import (
	"github.com/skycoin/cxo/skyobject"
	"reflect"
	"strings"
)

type wrappedObj struct {
	prev *wrappedObj
	next *wrappedObj

	s skyobject.SchemaReference
	p interface{}

	prevFieldName    string // Field name of prev obj used to find current.
	prevInFieldIndex int    // Index of prev obj's field's prevInFieldIndex. -1 if single reference (not array).

	w *RootWalker // Back reference.
}

func (w *RootWalker) newObj(s skyobject.SchemaReference, p interface{}, fn string, i int) *wrappedObj {
	return &wrappedObj{
		s:                s,
		p:                p,
		prevFieldName:    fn,
		prevInFieldIndex: i,
		w: w,
	}
}

func (o *wrappedObj) generate(s skyobject.SchemaReference, p interface{}, fn string, i int) *wrappedObj {
	newO := o.w.newObj(s, p, fn, i)
	newO.prev = o
	o.next = newO
	return newO
}

func (o *wrappedObj) elem() reflect.Value {
	return reflect.ValueOf(o.p).Elem()
}

func (o *wrappedObj) getFieldAsReferences(fieldName string) (
	refs skyobject.References, schemaName string, e error,
) {
	v := o.elem()
	vt := v.Type()

	// Obtain field.
	ft, has := vt.FieldByName(fieldName)
	if has == false {
		e = ErrFieldNotFound
		return
	}
	// Check type of field.
	if ft.Type.Kind() != reflect.Slice {
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

func (o *wrappedObj) getFieldAsReference(fieldName string) (
	ref skyobject.Reference, schemaName string, e error,
) {
	v := o.elem()
	vt := v.Type()

	// Obtain field.
	ft, has := vt.FieldByName(fieldName)
	if has == false {
		e = ErrFieldNotFound
		return
	}
	// Check type of field.
	if ft.Type.Kind() != reflect.Array {
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

func (o *wrappedObj) getFieldAsDynamic(fieldName string) (
	dyn skyobject.Dynamic, e error,
) {
	v := o.elem()
	vt := v.Type()

	// Obtain field.
	ft, has := vt.FieldByName(fieldName)
	if has == false {
		e = ErrFieldNotFound
		return
	}
	// Check type of field.
	if ft.Type.Kind() != reflect.Struct {
		e = ErrFieldHasWrongType
		return
	}

	// Obtain field value.
	f := v.FieldByName(fieldName)
	dyn = f.Interface().(skyobject.Dynamic)
	return
}

func (o *wrappedObj) getSchema(ct *skyobject.Container) skyobject.Schema {
	s, _ := ct.CoreRegistry().SchemaByReference(o.s)
	return s
}

func (o *wrappedObj) replaceReferencesField(fieldName string, newRefs skyobject.References) (e error) {
	v := o.elem()
	vt := v.Type()

	// Obtain field.
	ft, has := vt.FieldByName(fieldName)
	if has == false {
		e = ErrFieldNotFound
		return
	}
	// Check type of field.
	if ft.Type.Kind() != reflect.Slice {
		e = ErrFieldHasWrongType
		return
	}

	v.FieldByName(fieldName).Set(reflect.ValueOf(newRefs))
	return
}

func (o *wrappedObj) replaceReferenceField(fieldName string, newRef skyobject.Reference) (e error) {
	v := o.elem()
	vt := v.Type()

	// Obtain field.
	ft, has := vt.FieldByName(fieldName)
	if has == false {
		e = ErrFieldNotFound
		return
	}
	// Check type of field.
	if ft.Type.Kind() != reflect.Array {
		e = ErrFieldHasWrongType
		return
	}

	v.FieldByName(fieldName).Set(reflect.ValueOf(newRef))
	return
}

func (o *wrappedObj) replaceDynamicField(fieldName string, newDyn skyobject.Dynamic) (e error) {
	v := o.elem()
	vt := v.Type()

	// Obtain field.
	ft, has := vt.FieldByName(fieldName)
	if has == false {
		e = ErrFieldNotFound
		return
	}
	// Check type of field.
	if ft.Type.Kind() != reflect.Struct {
		e = ErrFieldHasWrongType
		return
	}

	v.FieldByName(fieldName).Set(reflect.ValueOf(newDyn))
	return
}

func (o *wrappedObj) save() (skyobject.Dynamic, error) {
	// Create dynamic reference of current object.
	dyn := skyobject.Dynamic{
		Object: o.w.c.Save(o.p),
		Schema: o.s,
	}

	// If this object is the direct child of root, save to root and return.
	if o.prev == nil {
		r := o.w.c.LastRoot(o.w.rpk)
		rDyns := r.Refs()
		rDyns[o.prevInFieldIndex] = dyn
		r.Replace(rDyns)
		return dyn, nil
	}

	// Get previous object's field type.
	v := o.prev.elem()
	vt := v.Type()

	sf, has := vt.FieldByName(o.prevFieldName)
	if has == false {
		return dyn, ErrFieldNotFound
	}

	switch sf.Type.Kind() {
	case reflect.Slice: // skyobject.References
		tRefs, _, e := o.prev.getFieldAsReferences(o.prevFieldName)
		if e != nil {
			return dyn, e
		}
		tRefs[o.prevInFieldIndex] = dyn.Object
		e = o.prev.replaceReferencesField(o.prevFieldName, tRefs)
		if e != nil {
			return dyn, e
		}
	case reflect.Array: // skyobject.Reference
		tRef, _, e := o.prev.getFieldAsReference(o.prevFieldName)
		if e != nil {
			return dyn, e
		}
		tRef = dyn.Object
		e = o.prev.replaceReferenceField(o.prevFieldName, tRef)
		if e != nil {
			return dyn, e
		}
	case reflect.Struct: // skyobject.Dynamic
		tDyn, e := o.prev.getFieldAsDynamic(o.prevFieldName)
		if e != nil {
			return dyn, e
		}
		tDyn = dyn
		e = o.prev.replaceDynamicField(o.prevFieldName, tDyn)
		if e != nil {
			return dyn, e
		}
	}

	return o.prev.save()
}

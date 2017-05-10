package skywalker

import (
	"github.com/skycoin/cxo/skyobject"
	"reflect"
	"strings"
)

type Obj struct {
	prev *Obj
	next *Obj

	p interface{}
	v *skyobject.Value

	fn    string
	index int // -1 if single reference (not array).
}

func NewObj(v *skyobject.Value, p interface{}) *Obj {
	return &Obj{
		p:     p,
		v:     v,
		index: -1,
	}
}

func (o *Obj) Generate(v *skyobject.Value, p interface{}, fn string, i int) *Obj {
	newO := NewObj(v, p)
	o.next = newO
	o.fn = fn
	o.index = i
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

//func Ha() {
//	obj := &Obj{}
//	//obj.value.
//}

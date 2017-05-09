package skywalker

import "github.com/skycoin/cxo/skyobject"

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

//func Ha() {
//	obj := &Obj{}
//	//obj.value.
//}

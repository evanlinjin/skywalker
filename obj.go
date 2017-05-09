package skywalker

import "github.com/skycoin/cxo/skyobject"

type Obj struct {
	prev *Obj
	next *Obj

	value *skyobject.Value

	refMember string
	refIndex  int // -1 if single reference (not array).
}

func NewObj(v *skyobject.Value) *Obj {
	return &Obj{
		value: v,
		refIndex: -1,
	}
}

//func Ha() {
//	obj := &Obj{}
//	//obj.value.
//}

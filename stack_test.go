package skywalker

import (
	"github.com/skycoin/cxo/skyobject"
	"github.com/skycoin/skycoin/src/cipher"
	"testing"
)

type Board struct {
	Name     string
	Creator  skyobject.Reference `skyobject:"schema=Person"`
	Featured skyobject.Dynamic
	Threads  skyobject.References `skyobject:"schema=Thread"`
}

type Thread struct {
	Name    string
	Creator skyobject.Reference `skyobject:"schema=Person"`
	Posts skyobject.References `skyobject:"schema=Post"`
}

type Post struct {
	Title string
	Body string
	Author skyobject.Reference `skyobject:"schema=Person"`
}

type Person struct {
	Name string
	Age uint64
}

// GENERATES:
// Public Key : 032ffee44b9554cd3350ee16760688b2fb9d0faae7f3534917ff07e971eb36fd6b
// Secret Key : b4f56cab07ea360c16c22ac241738e923b232138b69089fe0134f81a432ffaff
func genKeyPair() (cipher.PubKey, cipher.SecKey) {
	return cipher.GenerateDeterministicKeyPair([]byte("a"))
}

func newContainer() *skyobject.Container {
	r := skyobject.NewRegistry()
	r.Regsiter("Person", Person{})
	r.Regsiter("Post", Post{})
	r.Regsiter("Thread", Thread{})
	r.Regsiter("Board", Board{})
	r.Done()
	return skyobject.NewContainer(r)
}

func fillContainer1(c *skyobject.Container, pk cipher.PubKey, sk cipher.SecKey) {
	r := c.NewRoot(pk, sk)

	dynPerson := r.Dynamic(Person{"Dynamic Beast", 100})
	dynPost := r.Dynamic(Post{"Dynamic Post", "So big.", dynPerson.Object})

	persons := r.SaveArray(
		Person{"Evan", 21},
		Person{"Eric", 23},
		Person{"Jade", 24},
		Person{"Luis", 16},
	)
	posts1 := r.SaveArray(
		Post{"Hi", "Hello?", persons[0]},
		Post{"Bye", "Cya.", persons[0]},
		Post{"Howdy", "Haha.", persons[3]},
	)
	posts2 := r.SaveArray(
		Post{"OK", "Ok then...", persons[1]},
		Post{"What", "Eh what?", persons[2]},
		Post{"Is There?", "Is there really?", persons[3]},
	)
	posts3 := r.SaveArray(
		Post{"Test", "Yeah...", persons[2]},

	)
	threads := r.SaveArray(
		Thread{"Greetings", persons[0], posts1},
		Thread{"Expressions", persons[2], posts2},
		Thread{"Testing", persons[3], posts3},
	)
	r.InjectMany(
		Board{"Test", persons[3], dynPost, threads[2:]},
		Board{"Talk", persons[1], dynPerson, threads[:2]},
	)
}

func TestNewWalker(t *testing.T) {
	pk, sk := genKeyPair()
	c := newContainer()
	fillContainer1(c, pk, sk)
	_, e := NewWalker(c, pk, sk)
	if e != nil {
		t.Error("failed to create walker;", e)
	}
}

func TestWalker_AdvanceFromRoot(t *testing.T) {
	pk, sk := genKeyPair()
	c := newContainer()
	fillContainer1(c, pk, sk)
	w, _ := NewWalker(c, pk, sk)

	board := &Board{}
	e := w.AdvanceFromRoot(board, func(v *skyobject.Value) (chosen bool) {
		if v.Schema().Name() != "Board" {
			return false
		}
		fv, _ := v.FieldByName("Name")
		s, _ := fv.String()
		return s == "Talk"
	})
	if e != nil {
		t.Error("advance from root failed:", e)
	}
	t.Log(w.String())
}

func TestWalker_AdvanceFromRefsField(t *testing.T) {
	pk, sk := genKeyPair()
	c := newContainer()
	fillContainer1(c, pk, sk)
	w, _ := NewWalker(c, pk, sk)

	board := &Board{}
	thread := &Thread{}
	post := &Post{}

	var e error

	e = w.AdvanceFromRoot(board, func(v *skyobject.Value) (chosen bool) {
		if v.Schema().Name() != "Board" {
			return false
		}
		fv, _ := v.FieldByName("Name")
		s, _ := fv.String()
		return s == "Talk"
	})
	if e != nil {
		t.Error("advance from root failed:", e)
	}

	e = w.AdvanceFromRefsField("Threads", thread, func(v *skyobject.Value) (chosen bool) {
		fv, _ := v.FieldByName("Name")
		s, _ := fv.String()
		return s == "Greetings"
	})
	if e != nil {
		t.Error("advance from board failed:", e)
	}

	e = w.AdvanceFromRefsField("Posts", post, func(v *skyobject.Value) (chosen bool) {
		fv, _ := v.FieldByName("Title")
		s, _ := fv.String()
		return s == "Hi"
	})
	if e != nil {
		t.Error("advance from thread failed:", e)
	}
	t.Log("\n", w.String())
}
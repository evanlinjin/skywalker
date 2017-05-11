package skywalker

import (
	"github.com/skycoin/cxo/skyobject"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/cipher/encoder"
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
	Creator skyobject.Reference  `skyobject:"schema=Person"`
	Posts   skyobject.References `skyobject:"schema=Post"`
}

type Post struct {
	Title  string
	Body   string
	Author skyobject.Reference `skyobject:"schema=Person"`
}

type Person struct {
	Name string
	Age  uint64
}

// GENERATES:
// Public Key : 032ffee44b9554cd3350ee16760688b2fb9d0faae7f3534917ff07e971eb36fd6b
// Secret Key : b4f56cab07ea360c16c22ac241738e923b232138b69089fe0134f81a432ffaff
func genKeyPair() (cipher.PubKey, cipher.SecKey) {
	return cipher.GenerateDeterministicKeyPair([]byte("a"))
}

func newContainer() *skyobject.Container {
	r := skyobject.NewRegistry()
	r.Register("Person", Person{})
	r.Register("Post", Post{})
	r.Register("Thread", Thread{})
	r.Register("Board", Board{})
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
	_, e := NewRootWalker(c, pk, sk)
	if e != nil {
		t.Error("failed to create walker;", e)
	}
}

func TestWalker_AdvanceFromRoot(t *testing.T) {
	pk, sk := genKeyPair()
	c := newContainer()
	fillContainer1(c, pk, sk)
	w, _ := NewRootWalker(c, pk, sk)

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
	w, _ := NewRootWalker(c, pk, sk)

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
		t.Error("advance from root to board failed:", e)
	}

	e = w.AdvanceFromRefsField("Threads", thread, func(v *skyobject.Value) (chosen bool) {
		fv, _ := v.FieldByName("Name")
		s, _ := fv.String()
		return s == "Greetings"
	})
	if e != nil {
		t.Error("advance from board to thread failed:", e)
	}

	e = w.AdvanceFromRefsField("Posts", post, func(v *skyobject.Value) (chosen bool) {
		fv, _ := v.FieldByName("Title")
		s, _ := fv.String()
		return s == "Hi"
	})
	if e != nil {
		t.Error("advance from thread to post failed:", e)
	}
	t.Log("\n", w.String())
}

func TestWalker_AdvanceFromRefField(t *testing.T) {
	pk, sk := genKeyPair()
	c := newContainer()
	fillContainer1(c, pk, sk)
	w, _ := NewRootWalker(c, pk, sk)

	board := &Board{}
	thread := &Thread{}
	person := &Person{}

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
		t.Error("advance from root to board failed:", e)
	}

	e = w.AdvanceFromRefsField("Threads", thread, func(v *skyobject.Value) (chosen bool) {
		fv, _ := v.FieldByName("Name")
		s, _ := fv.String()
		return s == "Greetings"
	})
	if e != nil {
		t.Error("advance from board to thread failed:", e)
	}

	e = w.AdvanceFromRefField("Creator", person)
	if e != nil {
		t.Error("advance from thread to person failed:", e)
	}
	t.Log("\n", w.String())
}

func TestWalker_AdvanceFromDynamicField(t *testing.T) {
	t.Run("dynamic post", func(t *testing.T) {
		pk, sk := genKeyPair()
		c := newContainer()
		fillContainer1(c, pk, sk)
		w, _ := NewRootWalker(c, pk, sk)

		board := &Board{}
		post := &Post{}

		var e error

		e = w.AdvanceFromRoot(board, func(v *skyobject.Value) (chosen bool) {
			if v.Schema().Name() != "Board" {
				return false
			}
			fv, _ := v.FieldByName("Name")
			s, _ := fv.String()
			return s == "Test"
		})
		if e != nil {
			t.Error("advance from root to board failed:", e)
		}

		e = w.AdvanceFromDynamicField("Featured", post)
		if e != nil {
			t.Error("advance from board to dynamic post failed:", e)
		}
		t.Log("\n", w.String())
	})
	t.Run("dynamic person", func(t *testing.T) {
		pk, sk := genKeyPair()
		c := newContainer()
		fillContainer1(c, pk, sk)
		w, _ := NewRootWalker(c, pk, sk)

		board := &Board{}
		person := &Person{}

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
			t.Error("advance from root to board failed:", e)
		}

		e = w.AdvanceFromDynamicField("Featured", person)
		if e != nil {
			t.Error("advance from board to dynamic person failed:", e)
		}
		t.Log("\n", w.String())
	})
}

func TestWalker_AppendToRefsField(t *testing.T) {
	pk, sk := genKeyPair()
	c := newContainer()
	fillContainer1(c, pk, sk)
	w, _ := NewRootWalker(c, pk, sk)

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

	e = w.AppendToRefsField("Threads", Thread{Name: "New Thread"})
	if e != nil {
		t.Error("append thread to board failed:", e)
	}
	t.Log(w.String())

	thread := &Thread{}
	e = w.AdvanceFromRefsField("Threads", thread, func(v *skyobject.Value) (chosen bool) {
		fv, _ := v.FieldByName("Name")
		s, _ := fv.String()
		return s == "New Thread"
	})
	if e != nil {
		t.Error("advance from board to thread failed:", e)
	}
	t.Log(w.String())
}

func TestWalker_ReplaceInRefField(t *testing.T) {
	t.Run("depth of 1", func(t *testing.T) {
		pk, sk := genKeyPair()
		c := newContainer()
		fillContainer1(c, pk, sk)
		w, _ := NewRootWalker(c, pk, sk)

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

		e = w.ReplaceInRefField("Creator", Person{"Donald Trump", 70})
		if e != nil {
			t.Error("failed to replace:", e)
		}
		t.Log(w.String())
	})
	t.Run("depth of 2", func(t *testing.T) {
		pk, sk := genKeyPair()
		c := newContainer()
		fillContainer1(c, pk, sk)
		w, _ := NewRootWalker(c, pk, sk)

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

		thread := &Thread{}
		e = w.AdvanceFromRefsField("Threads", thread, func(v *skyobject.Value) (chosen bool) {
			fv, _ := v.FieldByName("Name")
			s, _ := fv.String()
			return s == "Greetings"
		})
		if e != nil {
			t.Error("advance from board to thread failed:", e)
		}

		t.Log(w.String())
		{
			p := &Person{}
			data, _ := c.Get(thread.Creator)
			encoder.DeserializeRaw(data, p)
			t.Log(p)
		}

		e = w.ReplaceInRefField("Creator", Person{Name: "Bruce Lee", Age: 77})
		if e != nil {
			t.Error("failed to replace", e)
		}

		t.Log(w.String())
		{
			p := &Person{}
			data, _ := c.Get(thread.Creator)
			encoder.DeserializeRaw(data, p)
			t.Log(p)
		}
	})
}

func TestWalker_ReplaceInDynamicField(t *testing.T) {
	t.Run("depth of 1", func(t *testing.T) {
		pk, sk := genKeyPair()
		c := newContainer()
		fillContainer1(c, pk, sk)
		w, _ := NewRootWalker(c, pk, sk)

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
		{
			p := &Person{}
			data, _ := c.Get(board.Featured.Object)
			encoder.DeserializeRaw(data, p)
			t.Log(p)
		}

		e = w.ReplaceInDynamicField("Featured", Post{Title: "Good Game", Body: "Yeah, this is fun."})
		if e != nil {
			t.Error("replace failed:", e)
		}

		t.Log(w.String())
		{
			p := &Post{}
			data, _ := c.Get(board.Featured.Object)
			encoder.DeserializeRaw(data, p)
			t.Log(p)
		}
	})
}
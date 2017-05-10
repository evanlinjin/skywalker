package skywalker

import (
	"github.com/skycoin/cxo/skyobject"
	"github.com/skycoin/skycoin/src/cipher"
	"testing"
)

type Board struct {
	Name  string
	Posts skyobject.References `skyobject:"schema=Post"`
}

type Post struct {
	Name string
}

// GENERATES:
// Public Key : 032ffee44b9554cd3350ee16760688b2fb9d0faae7f3534917ff07e971eb36fd6b
// Secret Key : b4f56cab07ea360c16c22ac241738e923b232138b69089fe0134f81a432ffaff
func genKeyPair() (cipher.PubKey, cipher.SecKey) {
	return cipher.GenerateDeterministicKeyPair([]byte("a"))
}

func newContainer() *skyobject.Container {
	r := skyobject.NewRegistry()
	r.Regsiter("Post", Post{})
	r.Regsiter("Board", Board{})
	r.Done()
	return skyobject.NewContainer(r)
}

func TestWalker_Advance(t *testing.T) {
	pk, sk := genKeyPair()
	c := newContainer()

	{
		r := c.NewRoot(pk, sk)
		pRefs := r.SaveArray(
			Post{Name: "Post 1"},
			Post{Name: "Post 2"},
			Post{Name: "Post 3"},
		)
		r.InjectMany(
			Board{Name: "Board 1"},
			Board{Name: "Board 2"},
			Board{Name: "Board 3", Posts: pRefs},
			Board{Name: "Board 4"},
		)
	}

	w, e := NewWalker(c, pk, sk)
	if e != nil {
		t.Error("failed;", e)
	}

	// OBTAIN THE BOARD >>>
	{
		board := &Board{}
		e = w.AdvanceFromRoot(board, func(v *skyobject.Value) bool {
			if v.Schema().Name() != "Board" {
				return false
			}
			fv, _ := v.FieldByName("Name")
			s, _ := fv.String()
			return s == "Board 3"
		})
		if e != nil {
			t.Error("failed to obtain board;", e)
		}
		t.Log("Obtained board:", *board)
		if board.Name != "Board 3" {
			t.Error("board name incorrect")
		}
		if w.Size() != 1 {
			t.Error("stack size incorrect")
		}
		if w.stack[0].p.(*Board).Name != "Board 3" {
			t.Error("interface board name incorrect")
		}
	}

	// OBTAIN THE POST >>>
	{
		post := &Post{}
		e = w.AdvanceFromRefsField("Posts", post, func(v *skyobject.Value) bool {
			fv, _ := v.FieldByName("Name")
			s, _ := fv.String()
			return s == "Post 2"
		})
		if e != nil {
			t.Error("failed to obtain post;", e)
		}
		t.Log("Obtained  post:", *post)
		if post.Name != "Post 2" {
			t.Error("post name incorrect")
		}
		if w.Size() != 2 {
			t.Error("stack size incorrect")
		}
		if w.stack[1].p.(*Post).Name != "Post 2" {
			t.Error("interface post name incorrect")
		}
	}
}

package skywalker

import (
	"github.com/skycoin/cxo/node"
	"github.com/skycoin/cxo/skyobject"
	"github.com/skycoin/skycoin/src/cipher"
	"testing"
	"time"
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

func makeClient() (*node.Client, error) {
	cc := node.NewClientConfig()
	c, e := node.NewClient(cc)
	if e != nil {
		return nil, e
	}
	if e := c.Start("[::]:8998"); e != nil {
		return nil, e
	}
	time.Sleep(5 * time.Second)
	c.Execute(func(ct *node.Container) error {
		ct.Register("Post", Post{})
		ct.Register("Board", Board{})
		return nil
	})
	return c, nil
}

func TestWalker_Advance(t *testing.T) {
	pk, sk := genKeyPair()

	c, e := makeClient()
	if e != nil {
		t.Error("failed to make cxo client;", e)
	}

	if c.Subscribe(pk) == false {
		t.Log("unable to subscribe to", pk.Hex())
	}

	c.Execute(func(ct *node.Container) error {
		r := ct.NewRoot(pk, sk)
		b1 := Board{Name: "Board 1"}
		b2 := Board{Name: "Board 2"}
		b3 := Board{Name: "Board 3"}
		b4 := Board{Name: "Board 4"}
		b5 := Board{Name: "Board 5"}
		r.Inject(b1, sk)
		r.Inject(b2, sk)
		r.Inject(b3, sk)
		r.Inject(b4, sk)
		r.Inject(b5, sk)
		return nil
	})

	e = c.Execute(func(ct *node.Container) error {
		w, e := NewWalker(ct, pk, sk)
		if e != nil {
			return e
		}
		b := &Board{}

		e = w.Advance(func(v *skyobject.Value) (found bool, e error) {
			fv, e := v.FieldByName("Name")
			if e != nil {
				return
			}
			s, e := fv.String()
			if e != nil {
				return
			}
			if s == "Board 3" {
				found = true
			}
			return
		}, b)
		if e != nil {
			return e
		}
		t.Log("Obtained board:", *b)
		return nil
	})
	if e != nil {
		t.Error("failed;", e)
	}
	c.Close()
}

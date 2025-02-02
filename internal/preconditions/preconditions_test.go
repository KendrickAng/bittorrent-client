package preconditions

import "testing"

func TestXor(t *testing.T) {
	if Xor(true, true) != false {
		t.Fatal()
	}
	if Xor(false, true) != true {
		t.Fatal()
	}
	if Xor(true, false) != true {
		t.Fatal()
	}
	if Xor(false, false) != false {
		t.Fatal()
	}
}

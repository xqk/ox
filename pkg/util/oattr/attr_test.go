package oattr

import "testing"

//
// TestNew
// @Description:
// @param t
//
func TestNew(t *testing.T) {
	k1 := 1
	v1 := "first"
	attr := New(k1, v1)

	ret1, ok1 := attr.Value(k1).(string)
	if !ok1 || v1 != ret1 {
		t.Fatalf("attr.Value error: want:%v ret:%v", v1, ret1)
	}

	k2 := "2"
	v2 := 2
	attr = attr.WithValues(k2, v2)
	ret2, ok2 := attr.Value(k2).(int)
	if !ok2 || v2 != ret2 {
		t.Fatalf("attr.WithValues error: want:%v ret:%v", v2, ret2)
	}
}
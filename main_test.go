package main

import "testing"

func TestSignature(t *testing.T) {
	got := "wrong"
	want := "right"

	if got != want {
		t.Errorf("got '%s' want '%s'", got, want)
	}
}

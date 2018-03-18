package main

import "testing"

func TestSignature(t *testing.T) {
	got := "right"
	want := "right"

	if got != want {
		t.Errorf("got '%s' want '%s'", got, want)
	}
}

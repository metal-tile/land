package main

import "testing"

func TestConvertXYToRowCol(t *testing.T) {
	col, row := ConvertXYToRowCol(968, 1004, 1)
	if e, g := 31, col; e != g {
		t.Fatalf("expected col is %d; got %d", e, g)
	}
	if e, g := 30, row; e != g {
		t.Fatalf("expected row is %d; got %d", e, g)
	}
}

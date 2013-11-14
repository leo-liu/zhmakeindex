package main

import (
	"testing"
)

func TestScanStyleTokens(t *testing.T) {
	data := []byte(" \t'hello'world")
	advance, token, err := ScanStyleTokens(data, false)
	if advance != 9 || string(token) != "'hello'" || err != nil {
		t.Fail()
	}
	advance0, token0, err0 := ScanStyleTokens(data[advance:], false)
	if advance0 != 0 || token0 != nil || err0 != nil {
		t.Fail()
	}
	advance1, token1, err1 := ScanStyleTokens(data[advance:], true)
	if advance1 != 5 || string(token1) != "world" || err1 != nil {
		t.Fail()
	}
}

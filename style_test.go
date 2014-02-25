package main

import (
	"testing"
)

func TestScanStyleTokens(t *testing.T) {
	data := []byte(" \t'hello'world")
	advance, token, err := ScanStyleTokens(data, false)
	if err != nil || advance != 9 || string(token) != "'hello'" {
		t.Error(err, string(token), advance)
	}
	advance0, token0, err0 := ScanStyleTokens(data[advance:], false)
	if err0 != nil || advance0 != 0 || token0 != nil {
		t.Error(err0, string(token0), advance0)
	}
	advance1, token1, err1 := ScanStyleTokens(data[advance:], true)
	if err1 != nil || advance1 != 5 || string(token1) != "world" {
		t.Error(err1, string(token1), advance1)
	}
}

func TestScanStyleTokens_comment(t *testing.T) {
	data := []byte("%comment\nfoo bar")
	adv, tok, err := ScanStyleTokens(data, false)
	if err != nil || string(tok) != "foo" || adv != 12 {
		t.Error(err, string(tok), adv)
	}

	data0 := []byte("\n%c\n\n%c\nfoo%c\nbar")
	adv0, tok0, err0 := ScanStyleTokens(data0, false)
	if err0 != nil || string(tok0) != "foo" || adv0 != 11 {
		t.Error(err0, string(tok0), adv0)
	}

	adv1, tok1, err1 := ScanStyleTokens(data0[adv0:], true)
	if err1 != nil || string(tok1) != "bar" || adv1 != 6 {
		t.Error(err1, string(tok1), adv1)
	}
}

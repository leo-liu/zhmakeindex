package main

import (
	"strings"
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

func TestScanStyleTokensBacktick(t *testing.T) {
	data := []byte("`hello` world")
	advance, token, err := ScanStyleTokens(data, false)
	if err != nil {
		t.Fatal(err)
	}
	if string(token) != "`hello`" {
		t.Fatalf("first token = %q, want %q", string(token), "`hello`")
	}
	if advance != len("`hello`") {
		t.Fatalf("advance = %d, want %d", advance, len("`hello`"))
	}
}

func TestUnquote(t *testing.T) {
	if got := unquote(`"hello"`); got != "hello" {
		t.Fatalf("unquote(\"\\\"hello\\\"\") = %q, want %q", got, "hello")
	}
	if got := unquote("`hello`"); got != "hello" {
		t.Fatalf("unquote(%q) = %q, want %q", "`hello`", got, "hello")
	}
}

func TestParseInt(t *testing.T) {
	if got := parseInt("42"); got != 42 {
		t.Fatalf("parseInt(%q) = %d, want %d", "42", got, 42)
	}
	if got := parseInt("0"); got != 0 {
		t.Fatalf("parseInt(%q) = %d, want %d", "0", got, 0)
	}
}

func TestNewInputStyle(t *testing.T) {
	style := NewInputStyle()
	if style.keyword != "\\indexentry" {
		t.Fatalf("keyword = %q, want %q", style.keyword, "\\indexentry")
	}
	if style.arg_open != '{' {
		t.Fatalf("arg_open = %q, want %q", style.arg_open, '{')
	}
	if style.level != '!' {
		t.Fatalf("level = %q, want %q", style.level, '!')
	}
	if style.comment != '%' {
		t.Fatalf("comment = %q, want %q", style.comment, '%')
	}
}

func TestNewOutputStyle(t *testing.T) {
	style := NewOutputStyle()
	if !strings.Contains(style.preamble, "\\begin{theindex}") {
		t.Fatalf("preamble = %q, want to contain %q", style.preamble, "\\begin{theindex}")
	}
	if style.delim_r != "--" {
		t.Fatalf("delim_r = %q, want %q", style.delim_r, "--")
	}
	if style.page_precedence != "rnaRA" {
		t.Fatalf("page_precedence = %q, want %q", style.page_precedence, "rnaRA")
	}
	if style.headings_flag != 0 {
		t.Fatalf("headings_flag = %d, want %d", style.headings_flag, 0)
	}
}

package main

import (
	"testing"

	"golang.org/x/text/encoding"
)

func TestStripExt(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "simple ext", input: "foo.idx", want: "foo"},
		{name: "path ext", input: "path/to/file.tex", want: "path/to/file"},
		{name: "no ext", input: "noext", want: "noext"},
		{name: "multiple dots", input: "a.b.c", want: "a.b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripExt(tt.input); got != tt.want {
				t.Fatalf("stripExt(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCheckEncoding(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, got encoding.Encoding)
	}{
		{
			name:  "utf-8 lowercase",
			input: "utf-8",
			check: func(t *testing.T, got encoding.Encoding) {
				if got != encoding.Nop {
					t.Fatalf("checkEncoding(%q) did not return encoding.Nop", "utf-8")
				}
			},
		},
		{
			name:  "utf-8 uppercase",
			input: "UTF-8",
			check: func(t *testing.T, got encoding.Encoding) {
				if got != encoding.Nop {
					t.Fatalf("checkEncoding(%q) did not return encoding.Nop", "UTF-8")
				}
			},
		},
		{
			name:  "gbk",
			input: "gbk",
			check: func(t *testing.T, got encoding.Encoding) {
				if got == nil {
					t.Fatalf("checkEncoding(%q) returned nil", "gbk")
				}
			},
		},
		{
			name:  "big5",
			input: "big5",
			check: func(t *testing.T, got encoding.Encoding) {
				if got == nil {
					t.Fatalf("checkEncoding(%q) returned nil", "big5")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkEncoding(tt.input)
			tt.check(t, got)
		})
	}
}

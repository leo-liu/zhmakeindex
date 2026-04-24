package main

import (
	"io"
	"strings"
	"testing"
)

func TestNumberdReaderBasic(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantRunes    []rune
		wantLines    []int
		finalLineNum int
	}{
		{
			name:         "basic read with newline",
			input:        "abc\ndef",
			wantRunes:    []rune{'a', 'b', 'c', '\n', 'd'},
			wantLines:    []int{1, 1, 1, 2, 2},
			finalLineNum: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewNumberdReader(strings.NewReader(tt.input))

			for i, wantRune := range tt.wantRunes {
				gotRune, _, err := reader.ReadRune()
				if err != nil {
					t.Fatalf("ReadRune() error at step %d: %v", i, err)
				}
				if gotRune != wantRune {
					t.Fatalf("ReadRune() rune at step %d = %q, want %q", i, gotRune, wantRune)
				}
				if gotLine := reader.Line(); gotLine != tt.wantLines[i] {
					t.Fatalf("Line() after step %d = %d, want %d", i, gotLine, tt.wantLines[i])
				}
			}

			if got := reader.Line(); got != tt.finalLineNum {
				t.Fatalf("final Line() = %d, want %d", got, tt.finalLineNum)
			}
		})
	}
}

func TestNumberdReaderUnreadRune(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "unread regular rune and newline",
			input: "a\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewNumberdReader(strings.NewReader(tt.input))

			gotRune, _, err := reader.ReadRune()
			if err != nil {
				t.Fatalf("first ReadRune() error: %v", err)
			}
			if gotRune != 'a' {
				t.Fatalf("first ReadRune() = %q, want %q", gotRune, 'a')
			}
			if gotLine := reader.Line(); gotLine != 1 {
				t.Fatalf("Line() after reading 'a' = %d, want 1", gotLine)
			}

			if err := reader.UnreadRune(); err != nil {
				t.Fatalf("UnreadRune() after 'a' error: %v", err)
			}
			if gotLine := reader.Line(); gotLine != 1 {
				t.Fatalf("Line() after unreading 'a' = %d, want 1", gotLine)
			}

			gotRune, _, err = reader.ReadRune()
			if err != nil {
				t.Fatalf("second ReadRune() error: %v", err)
			}
			if gotRune != 'a' {
				t.Fatalf("second ReadRune() = %q, want %q", gotRune, 'a')
			}
			if gotLine := reader.Line(); gotLine != 1 {
				t.Fatalf("Line() after rereading 'a' = %d, want 1", gotLine)
			}

			gotRune, _, err = reader.ReadRune()
			if err != nil {
				t.Fatalf("third ReadRune() error: %v", err)
			}
			if gotRune != '\n' {
				t.Fatalf("third ReadRune() = %q, want %q", gotRune, '\n')
			}
			if gotLine := reader.Line(); gotLine != 2 {
				t.Fatalf("Line() after reading newline = %d, want 2", gotLine)
			}

			if err := reader.UnreadRune(); err != nil {
				t.Fatalf("UnreadRune() after newline error: %v", err)
			}
			if gotLine := reader.Line(); gotLine != 1 {
				t.Fatalf("Line() after unreading newline = %d, want 1", gotLine)
			}

			gotRune, _, err = reader.ReadRune()
			if err != nil {
				t.Fatalf("fourth ReadRune() error: %v", err)
			}
			if gotRune != '\n' {
				t.Fatalf("fourth ReadRune() = %q, want %q", gotRune, '\n')
			}
			if gotLine := reader.Line(); gotLine != 2 {
				t.Fatalf("Line() after rereading newline = %d, want 2", gotLine)
			}
		})
	}
}

func TestNumberdReaderSkipLine(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		wantFirstRune      rune
		wantLineAfterFirst int
		wantNextRune       rune
		wantLineAfterSkip  int
	}{
		{
			name:               "skip remainder of current line",
			input:              "abc\ndef\nghi",
			wantFirstRune:      'a',
			wantLineAfterFirst: 1,
			wantNextRune:       'd',
			wantLineAfterSkip:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewNumberdReader(strings.NewReader(tt.input))

			gotRune, _, err := reader.ReadRune()
			if err != nil {
				t.Fatalf("ReadRune() before SkipLine error: %v", err)
			}
			if gotRune != tt.wantFirstRune {
				t.Fatalf("ReadRune() before SkipLine = %q, want %q", gotRune, tt.wantFirstRune)
			}
			if gotLine := reader.Line(); gotLine != tt.wantLineAfterFirst {
				t.Fatalf("Line() after first rune = %d, want %d", gotLine, tt.wantLineAfterFirst)
			}

			if err := reader.SkipLine(); err != nil {
				t.Fatalf("SkipLine() error: %v", err)
			}
			if gotLine := reader.Line(); gotLine != tt.wantLineAfterSkip {
				t.Fatalf("Line() after SkipLine = %d, want %d", gotLine, tt.wantLineAfterSkip)
			}

			gotRune, _, err = reader.ReadRune()
			if err != nil {
				t.Fatalf("ReadRune() after SkipLine error: %v", err)
			}
			if gotRune != tt.wantNextRune {
				t.Fatalf("ReadRune() after SkipLine = %q, want %q", gotRune, tt.wantNextRune)
			}
			if gotLine := reader.Line(); gotLine != tt.wantLineAfterSkip {
				t.Fatalf("Line() after reading next rune = %d, want %d", gotLine, tt.wantLineAfterSkip)
			}
		})
	}
}

func TestNumberdReaderUnicode(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantRunes []rune
		wantLines []int
	}{
		{
			name:      "unicode runes are read correctly",
			input:     "你好\n世界",
			wantRunes: []rune{'你', '好', '\n', '世'},
			wantLines: []int{1, 1, 2, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewNumberdReader(strings.NewReader(tt.input))

			for i, wantRune := range tt.wantRunes {
				gotRune, _, err := reader.ReadRune()
				if err != nil {
					t.Fatalf("ReadRune() error at step %d: %v", i, err)
				}
				if gotRune != wantRune {
					t.Fatalf("ReadRune() rune at step %d = %q, want %q", i, gotRune, wantRune)
				}
				if gotLine := reader.Line(); gotLine != tt.wantLines[i] {
					t.Fatalf("Line() after step %d = %d, want %d", i, gotLine, tt.wantLines[i])
				}
			}
		})
	}
}

func TestNumberdReaderEmpty(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty input returns eof",
			input: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewNumberdReader(strings.NewReader(tt.input))

			_, _, err := reader.ReadRune()
			if err != io.EOF {
				t.Fatalf("ReadRune() error = %v, want %v", err, io.EOF)
			}
			if gotLine := reader.Line(); gotLine != 1 {
				t.Fatalf("Line() after EOF = %d, want 1", gotLine)
			}
		})
	}
}

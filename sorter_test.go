package main

import "testing"

func TestIsNumRune(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want bool
	}{
		{name: "zero", r: '0', want: true},
		{name: "nine", r: '9', want: true},
		{name: "circled one", r: '①', want: true},
		{name: "special zero", r: '〇', want: false},
		{name: "latin", r: 'a', want: false},
		{name: "cjk", r: '中', want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNumRune(tt.r); got != tt.want {
				t.Fatalf("IsNumRune(%q) = %v, want %v", tt.r, got, tt.want)
			}
		})
	}
}

func TestIsNumString(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{name: "digits", s: "123", want: true},
		{name: "single zero", s: "0", want: true},
		{name: "empty", s: "", want: true},
		{name: "mixed", s: "12a", want: false},
		{name: "special zero", s: "〇", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNumString(tt.s); got != tt.want {
				t.Fatalf("IsNumString(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestDecimalStrcmp(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{name: "two less than ten", a: "2", b: "10", want: -1},
		{name: "ten greater than two", a: "10", b: "2", want: 1},
		{name: "same number", a: "5", b: "5", want: 0},
		{name: "non numeric equal", a: "abc", b: "def", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DecimalStrcmp(tt.a, tt.b); got != tt.want {
				t.Fatalf("DecimalStrcmp(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestRuneCmpIgnoreCases(t *testing.T) {
	tests := []struct {
		name string
		a    rune
		b    rune
		check func(int) bool
	}{
		{name: "same letter different case", a: 'a', b: 'A', check: func(got int) bool { return got == 0 }},
		{name: "a before b", a: 'a', b: 'b', check: func(got int) bool { return got < 0 }},
		{name: "B after a", a: 'B', b: 'a', check: func(got int) bool { return got > 0 }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RuneCmpIgnoreCases(tt.a, tt.b)
			if !tt.check(got) {
				t.Fatalf("RuneCmpIgnoreCases(%q, %q) = %d", tt.a, tt.b, got)
			}
		})
	}
}

func TestGetStringType(t *testing.T) {
	collator := ReadingIndexCollator{}
	tests := []struct {
		name string
		s    string
		want stringType
	}{
		{name: "empty", s: "", want: EMPTY_STR},
		{name: "symbol", s: "+", want: SYMBOL_STR},
		{name: "numeric", s: "123", want: NUM_STR},
		{name: "num symbol", s: "12abc", want: NUM_SYMBOL_STR},
		{name: "letter", s: "hello", want: LETTER_STR},
		{name: "cjk letter", s: "中文", want: LETTER_STR},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getStringType(collator, tt.s); got != tt.want {
				t.Fatalf("getStringType(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestIndexEntrySliceStrcmp(t *testing.T) {
	s := IndexEntrySlice{colattor: ReadingIndexCollator{}}
	tests := []struct {
		name  string
		a     string
		b     string
		check func(int) bool
	}{
		{name: "abc before def", a: "abc", b: "def", check: func(got int) bool { return got < 0 }},
		{name: "same string", a: "abc", b: "abc", check: func(got int) bool { return got == 0 }},
		{name: "case fallback to raw string", a: "ABC", b: "abc", check: func(got int) bool { return got == -1 }},
		{name: "natural number compare", a: "2", b: "10", check: func(got int) bool { return got < 0 }},
		{name: "empty string smallest", a: "", b: "abc", check: func(got int) bool { return got < 0 }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.Strcmp(tt.a, tt.b)
			if !tt.check(got) {
				t.Fatalf("Strcmp(%q, %q) = %d", tt.a, tt.b, got)
			}
		})
	}
}

func TestNewPageSorterPrecedence(t *testing.T) {
	tests := []struct {
		name    string
		style   *OutputStyle
		option  *OutputOptions
		format  NumFormat
		wantPos int
	}{
		{
			name:    "default roman lower precedence first",
			style:   NewOutputStyle(),
			option:  &OutputOptions{},
			format:  NUM_ROMAN_LOWER,
			wantPos: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sorter := NewPageSorter(tt.style, tt.option)
			if got := sorter.precedence[tt.format]; got != tt.wantPos {
				t.Fatalf("precedence[%v] = %d, want %d", tt.format, got, tt.wantPos)
			}
		})
	}
}

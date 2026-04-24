package main

import (
	"testing"
)

func TestScanNumber(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		want    PageNumber
		wantErr error
	}{
		{name: "arabic one", token: "1", want: PageNumber{format: NUM_ARABIC, num: 1}},
		{name: "arabic forty two", token: "42", want: PageNumber{format: NUM_ARABIC, num: 42}},
		{name: "arabic zero", token: "0", want: PageNumber{format: NUM_ARABIC, num: 0}},
		{name: "roman lower one", token: "i", want: PageNumber{format: NUM_ROMAN_LOWER, num: 1}},
		{name: "roman lower four", token: "iv", want: PageNumber{format: NUM_ROMAN_LOWER, num: 4}},
		{name: "roman lower fourteen", token: "xiv", want: PageNumber{format: NUM_ROMAN_LOWER, num: 14}},
		{name: "roman upper three", token: "III", want: PageNumber{format: NUM_ROMAN_UPPER, num: 3}},
		{name: "roman upper fourteen", token: "XIV", want: PageNumber{format: NUM_ROMAN_UPPER, num: 14}},
		{name: "alpha lower a", token: "a", want: PageNumber{format: NUM_ALPH_LOWER, num: 1}},
		{name: "alpha lower z", token: "z", want: PageNumber{format: NUM_ALPH_LOWER, num: 26}},
		{name: "alpha upper A", token: "A", want: PageNumber{format: NUM_ALPH_UPPER, num: 1}},
		{name: "empty token", token: "", wantErr: ScanSyntaxError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := scanNumber([]rune(tt.token))
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("scanNumber(%q) error = %v, want %v", tt.token, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("scanNumber(%q) unexpected error: %v", tt.token, err)
			}
			if got != tt.want {
				t.Fatalf("scanNumber(%q) = %+v, want %+v", tt.token, got, tt.want)
			}
		})
	}
}

func TestScanPage(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		compositor string
		want       []PageNumber
		wantErr    error
	}{
		{
			name:       "composite page",
			token:      "5-ii-1",
			compositor: "-",
			want: []PageNumber{
				{format: NUM_ARABIC, num: 5},
				{format: NUM_ROMAN_LOWER, num: 2},
				{format: NUM_ARABIC, num: 1},
			},
		},
		{
			name:       "single page",
			token:      "42",
			compositor: "-",
			want: []PageNumber{{format: NUM_ARABIC, num: 42}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := scanPage([]rune(tt.token), tt.compositor)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("scanPage(%q, %q) error = %v, want %v", tt.token, tt.compositor, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("scanPage(%q, %q) unexpected error: %v", tt.token, tt.compositor, err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("scanPage(%q, %q) len = %d, want %d", tt.token, tt.compositor, len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("scanPage(%q, %q)[%d] = %+v, want %+v", tt.token, tt.compositor, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestRomanNumString(t *testing.T) {
	tests := []struct {
		name  string
		num   int
		upper bool
		want  string
	}{
		{name: "lower 1", num: 1, want: "i"},
		{name: "lower 4", num: 4, want: "iv"},
		{name: "lower 9", num: 9, want: "ix"},
		{name: "lower 14", num: 14, want: "xiv"},
		{name: "lower 1999", num: 1999, want: "mcmxcix"},
		{name: "upper 1", num: 1, upper: true, want: "I"},
		{name: "upper 14", num: 14, upper: true, want: "XIV"},
		{name: "zero", num: 0, want: ""},
		{name: "negative", num: -1, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := romanNumString(tt.num, tt.upper)
			if got != tt.want {
				t.Fatalf("romanNumString(%d, %v) = %q, want %q", tt.num, tt.upper, got, tt.want)
			}
		})
	}
}

func TestNumFormatFormat(t *testing.T) {
	tests := []struct {
		name   string
		format NumFormat
		num    int
		want   string
	}{
		{name: "arabic", format: NUM_ARABIC, num: 42, want: "42"},
		{name: "roman lower", format: NUM_ROMAN_LOWER, num: 4, want: "iv"},
		{name: "alpha lower", format: NUM_ALPH_LOWER, num: 1, want: "b"},
		{name: "alpha upper", format: NUM_ALPH_UPPER, num: 1, want: "B"},
		{name: "unknown", format: NUM_UNKNOWN, num: 0, want: "?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.format.Format(tt.num)
			if got != tt.want {
				t.Fatalf("%v.Format(%d) = %q, want %q", tt.format, tt.num, got, tt.want)
			}
		})
	}
}

func TestPageString(t *testing.T) {
	page := &Page{
		numbers: []PageNumber{
			{format: NUM_ARABIC, num: 5},
			{format: NUM_ROMAN_LOWER, num: 2},
			{format: NUM_ARABIC, num: 1},
		},
		compositor: "-",
	}

	got := page.String()
	want := "5-ii-1"
	if got != want {
		t.Fatalf("Page.String() = %q, want %q", got, want)
	}
}

func TestPageCompatible(t *testing.T) {
	tests := []struct {
		name  string
		page  *Page
		other *Page
		want  bool
	}{
		{
			name: "same formats are compatible",
			page: &Page{numbers: []PageNumber{{format: NUM_ARABIC, num: 5}, {format: NUM_ROMAN_LOWER, num: 2}}},
			other: &Page{numbers: []PageNumber{{format: NUM_ARABIC, num: 8}, {format: NUM_ROMAN_LOWER, num: 4}}},
			want: true,
		},
		{
			name: "different formats are incompatible",
			page: &Page{numbers: []PageNumber{{format: NUM_ARABIC, num: 5}, {format: NUM_ROMAN_LOWER, num: 2}}},
			other: &Page{numbers: []PageNumber{{format: NUM_ARABIC, num: 5}, {format: NUM_ARABIC, num: 2}}},
			want: false,
		},
		{
			name: "different lengths are incompatible",
			page: &Page{numbers: []PageNumber{{format: NUM_ARABIC, num: 42}}},
			other: &Page{numbers: []PageNumber{{format: NUM_ARABIC, num: 42}, {format: NUM_ROMAN_LOWER, num: 1}}},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.page.Compatible(tt.other)
			if got != tt.want {
				t.Fatalf("Compatible() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPageDiff(t *testing.T) {
	tests := []struct {
		name  string
		page  *Page
		other *Page
		want  int
	}{
		{
			name:  "compatible last segment differs",
			page:  &Page{numbers: []PageNumber{{format: NUM_ARABIC, num: 5}, {format: NUM_ROMAN_LOWER, num: 2}, {format: NUM_ARABIC, num: 1}}},
			other: &Page{numbers: []PageNumber{{format: NUM_ARABIC, num: 5}, {format: NUM_ROMAN_LOWER, num: 2}, {format: NUM_ARABIC, num: 4}}},
			want:  3,
		},
		{
			name:  "incompatible returns minus one",
			page:  &Page{numbers: []PageNumber{{format: NUM_ARABIC, num: 5}}},
			other: &Page{numbers: []PageNumber{{format: NUM_ROMAN_LOWER, num: 5}}},
			want:  -1,
		},
		{
			name:  "non last segment differs returns maxint",
			page:  &Page{numbers: []PageNumber{{format: NUM_ARABIC, num: 5}, {format: NUM_ROMAN_LOWER, num: 2}, {format: NUM_ARABIC, num: 1}}},
			other: &Page{numbers: []PageNumber{{format: NUM_ARABIC, num: 6}, {format: NUM_ROMAN_LOWER, num: 2}, {format: NUM_ARABIC, num: 1}}},
			want:  MaxInt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.page.Diff(tt.other)
			if got != tt.want {
				t.Fatalf("Diff() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPageCmp(t *testing.T) {
	precedence := map[NumFormat]int{
		NUM_ROMAN_LOWER: 0,
		NUM_ARABIC:      1,
		NUM_ALPH_LOWER:  2,
		NUM_ROMAN_UPPER: 3,
		NUM_ALPH_UPPER:  4,
	}

	tests := []struct {
		name  string
		page  *Page
		other *Page
		want  int
	}{
		{
			name:  "equal pages compare equal",
			page:  &Page{numbers: []PageNumber{{format: NUM_ARABIC, num: 42}}},
			other: &Page{numbers: []PageNumber{{format: NUM_ARABIC, num: 42}}},
			want:  0,
		},
		{
			name:  "precedence decides format order",
			page:  &Page{numbers: []PageNumber{{format: NUM_ROMAN_LOWER, num: 1}}},
			other: &Page{numbers: []PageNumber{{format: NUM_ARABIC, num: 1}}},
			want:  -1,
		},
		{
			name:  "number decides within same format",
			page:  &Page{numbers: []PageNumber{{format: NUM_ARABIC, num: 5}}},
			other: &Page{numbers: []PageNumber{{format: NUM_ARABIC, num: 7}}},
			want:  -2,
		},
		{
			name:  "prefix equal shorter page is smaller",
			page:  &Page{numbers: []PageNumber{{format: NUM_ARABIC, num: 5}}},
			other: &Page{numbers: []PageNumber{{format: NUM_ARABIC, num: 5}, {format: NUM_ROMAN_LOWER, num: 1}}},
			want:  -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.page.Cmp(tt.other, precedence)
			if got != tt.want {
				t.Fatalf("Cmp() = %d, want %d", got, tt.want)
			}
		})
	}
}

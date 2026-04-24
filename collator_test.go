package main

import "testing"

func TestReadingCollatorIsLetter(t *testing.T) {
	collator := ReadingIndexCollator{}

	tests := []struct {
		name string
		r    rune
		want bool
	}{
		{name: "lower ascii", r: 'a', want: true},
		{name: "upper ascii", r: 'Z', want: true},
		{name: "cjk with reading", r: '中', want: true},
		{name: "symbol", r: '+', want: false},
		{name: "digit", r: '0', want: false},
	}

	for _, tt := range tests {
		if got := collator.IsLetter(tt.r); got != tt.want {
			t.Fatalf("IsLetter(%q) = %v, want %v", tt.r, got, tt.want)
		}
	}
}

func TestReadingCollatorRuneCmp(t *testing.T) {
	collator := ReadingIndexCollator{}

	if got := collator.RuneCmp('a', 'b'); got >= 0 {
		t.Fatalf("RuneCmp('a', 'b') = %d, want < 0", got)
	}
	if got := collator.RuneCmp('a', 'A'); got != 0 {
		t.Fatalf("RuneCmp('a', 'A') = %d, want 0", got)
	}
	if got := collator.RuneCmp('中', '文'); got <= 0 {
		t.Fatalf("RuneCmp('中', '文') = %d, want > 0 because zhong1 > wen2", got)
	}
	if got := collator.RuneCmp('文', '中'); got >= 0 {
		t.Fatalf("RuneCmp('文', '中') = %d, want < 0 because wen2 < zhong1", got)
	}
}

func TestReadingCollatorGroup(t *testing.T) {
	collator := ReadingIndexCollator{}

	letterEntry := &IndexEntry{
		level: []IndexEntryLevel{{key: "abc", text: "abc"}},
	}
	if got := collator.Group(letterEntry); got != 2 {
		t.Fatalf("Group(letterEntry) = %d, want 2", got)
	}

	numberEntry := &IndexEntry{
		level: []IndexEntryLevel{{key: "123", text: "123"}},
	}
	if got := collator.Group(numberEntry); got != 1 {
		t.Fatalf("Group(numberEntry) = %d, want 1", got)
	}
}

func TestReadingCollatorInitGroups(t *testing.T) {
	collator := ReadingIndexCollator{}
	style := NewOutputStyle()

	groups := collator.InitGroups(style)
	if len(groups) != 28 {
		t.Fatalf("len(InitGroups()) = %d, want 28", len(groups))
	}
}

func TestStrokeCollatorIsLetter(t *testing.T) {
	collator := StrokeIndexCollator{}

	tests := []struct {
		r    rune
		want bool
	}{
		{r: 'a', want: true},
		{r: '中', want: true},
		{r: '+', want: false},
	}

	for _, tt := range tests {
		if got := collator.IsLetter(tt.r); got != tt.want {
			t.Fatalf("IsLetter(%q) = %v, want %v", tt.r, got, tt.want)
		}
	}
}

func TestStrokeCollatorRuneCmp(t *testing.T) {
	collator := StrokeIndexCollator{}

	if got := collator.RuneCmp('一', '中'); got >= 0 {
		t.Fatalf("RuneCmp('一', '中') = %d, want < 0 because 1 stroke < 4 strokes", got)
	}
}

func TestRadicalCollatorIsLetter(t *testing.T) {
	collator := RadicalIndexCollator{}

	tests := []struct {
		r    rune
		want bool
	}{
		{r: 'a', want: true},
		{r: '中', want: true},
		{r: '+', want: false},
	}

	for _, tt := range tests {
		if got := collator.IsLetter(tt.r); got != tt.want {
			t.Fatalf("IsLetter(%q) = %v, want %v", tt.r, got, tt.want)
		}
	}
}

func TestRadicalCollatorRuneCmp(t *testing.T) {
	collator := RadicalIndexCollator{}

	if got := collator.RuneCmp('中', '文'); got >= 0 {
		t.Fatalf("RuneCmp('中', '文') = %d, want < 0 because radical data orders 中 before 文", got)
	}
	if got := collator.RuneCmp('文', '中'); got <= 0 {
		t.Fatalf("RuneCmp('文', '中') = %d, want > 0 because radical data orders 文 after 中", got)
	}
}

func TestCollatorIsLetterConsistency(t *testing.T) {
	collators := []struct {
		name string
		c    IndexCollator
	}{
		{name: "reading", c: ReadingIndexCollator{}},
		{name: "stroke", c: StrokeIndexCollator{}},
		{name: "radical", c: RadicalIndexCollator{}},
	}

	runes := []rune{'a', 'Z', '+', '0', '中'}
	for _, r := range runes {
		base := collators[0].c.IsLetter(r)
		for _, collator := range collators[1:] {
			if got := collator.c.IsLetter(r); got != base {
				t.Fatalf("IsLetter(%q) mismatch: %s=%v, %s=%v", r, collators[0].name, base, collator.name, got)
			}
		}
	}
}

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "zhmakeindex")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("build failed: %v", err)
	}
	return bin
}

func normalizeLineEndings(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}

func TestIntegration(t *testing.T) {
	bin := buildBinary(t)

	idxFiles := []string{"numbers", "symorder", "mixedpage", "rangeencap", "compositepage"}
	sortMethods := []string{"pinyin", "stroke", "radical"}

	for _, f := range idxFiles {
		for _, z := range sortMethods {
			name := f + "_" + z
			t.Run(name, func(t *testing.T) {
				goldenPath := filepath.Join("testdata", name+".golden")
				golden, err := os.ReadFile(goldenPath)
				if err != nil {
					t.Fatalf("read golden: %v", err)
				}

				tmpDir := t.TempDir()
				outPath := filepath.Join(tmpDir, name+".ind")
				logPath := filepath.Join(tmpDir, name+".ilg")
				cmd := exec.Command(bin, "-z", z, "-o", outPath, "-t", logPath, filepath.Join("examples", f+".idx"))
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					t.Fatalf("run failed: %v", err)
				}

				got, err := os.ReadFile(outPath)
				if err != nil {
					t.Fatalf("read output: %v", err)
				}
				if normalizeLineEndings(string(got)) != normalizeLineEndings(string(golden)) {
					t.Errorf("output mismatch for %s\n--- golden len=%d\n+++ got len=%d", name, len(golden), len(got))
				}
			})
		}
	}
}

func TestIntegrationWithStyle(t *testing.T) {
	bin := buildBinary(t)

	tests := []struct {
		name  string
		style string
		idx   string
	}{
		{name: "zh_ist", style: "examples/zh.ist", idx: "examples/numbers.idx"},
		{name: "suffix_ist", style: "examples/suffix.ist", idx: "examples/numbers.idx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goldenPath := filepath.Join("testdata", "numbers_"+tt.name+".golden")
			golden, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("read golden: %v", err)
			}

			tmpDir := t.TempDir()
			outPath := filepath.Join(tmpDir, tt.name+".ind")
			logPath := filepath.Join(tmpDir, tt.name+".ilg")
			cmd := exec.Command(bin, "-s", tt.style, "-z", "pinyin", "-o", outPath, "-t", logPath, tt.idx)
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				t.Fatalf("run failed: %v", err)
			}

			got, err := os.ReadFile(outPath)
			if err != nil {
				t.Fatalf("read output: %v", err)
			}
			if normalizeLineEndings(string(got)) != normalizeLineEndings(string(golden)) {
				t.Errorf("output mismatch for %s\n--- golden len=%d\n+++ got len=%d", tt.name, len(golden), len(got))
			}
		})
	}
}

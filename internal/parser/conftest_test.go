package parser

import (
	"path/filepath"
	"testing"
)

func TestParseConftestSARIF(t *testing.T) {
	path := filepath.Join("..", "..", "fixtures", "conftest.sarif.json")
	f, err := ParseConftest(path)
	if err != nil {
		t.Fatal(err)
	}
	if f.High != 1 || f.Critical != 0 {
		t.Fatalf("got critical=%d high=%d", f.Critical, f.High)
	}
}

func TestParseConftestEmpty(t *testing.T) {
	f, err := ParseConftest("")
	if err != nil || f.Critical != 0 || f.High != 0 {
		t.Fatalf("empty: %+v err=%v", f, err)
	}
}

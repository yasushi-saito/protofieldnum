package main

import (
	"os"
	"path/filepath"
	"testing"

	protoparser "github.com/yoheimuta/go-protoparser/v4"
)

func TestGetNextID(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	path := filepath.Join(wd, "test.proto")
	fd, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open %v: %v", path, err)
	}
	defer fd.Close()

	parsed, err := protoparser.Parse(fd,
		protoparser.WithDebug(false),
		protoparser.WithFilename(filepath.Base(path)))
	if err != nil {
		t.Fatalf("Parse %v: %v", path, err)
	}
	pp, err := protoparser.UnorderedInterpret(parsed)
	if err != nil {
		t.Fatalf("Interpret %v: %v", path, err)
	}

	tests := []struct {
		msgName string
		want    int
	}{
		{"M0", 2},
		{"M1", 2}, // First M1
		{"M2", 3}, // Second M1
		{"M3", 2},
		{"M4", 11}, // reserved 2, 10
		{"M5", 2},
	}

	for i, m := range pp.ProtoBody.Messages {
		tc := tests[i]
		t.Run(tc.msgName, func(t *testing.T) {
			if m.MessageName != tc.msgName {
				t.Errorf("MessageName mismatch: got %s, want %s", m.MessageName, tc.msgName)
			}
			got := getNextID(m)
			if got != tc.want {
				t.Errorf("getNextID() = %d, want %d", got, tc.want)
			}
		})
	}
}

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	protoparser "github.com/yoheimuta/go-protoparser/v4"
	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"
)

func mustParseFieldNumber(s string, fieldPath string) int {
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		log.Fatalf("%s: Could not parse field number '%s': %v", fieldPath, s, err)
	}
	return int(n)
}

func getNextID(m *unordered.Message) int {
	used := map[int]bool{}
	for _, f := range m.MessageBody.Fields {
		n := mustParseFieldNumber(f.FieldNumber, fmt.Sprintf("%s.%s", m.MessageName, f.FieldName))
		used[n] = true
	}
	for _, mp := range m.MessageBody.Maps {
		n := mustParseFieldNumber(mp.FieldNumber, fmt.Sprintf("%s.%s", m.MessageName, mp.MapName))
		used[n] = true
	}
	for _, oneof := range m.MessageBody.Oneofs {
		for _, f := range oneof.OneofFields {
			n := mustParseFieldNumber(f.FieldNumber, fmt.Sprintf("%s.%s", m.MessageName, f.FieldName))
			used[n] = true
		}
	}
	for _, r := range m.MessageBody.Reserves {
		for _, rg := range r.Ranges {
			start := mustParseFieldNumber(rg.Begin, fmt.Sprintf("%s(reserved) %v", m.MessageName, rg.Begin))
			var end int
			if rg.End == "" {
				end = start
			} else {
				end = mustParseFieldNumber(rg.End, fmt.Sprintf("%s(reserved) %v", m.MessageName, rg.End))
			}
			if end < start {
				log.Fatalf("%s %v: end <= start", m.MessageName, rg)
			}
			for i := start; i <= end; i++ {
				used[i] = true
			}
		}
	}
	for i := 1; i < 1000000; i++ {
		if _, ok := used[i]; !ok {
			return i
		}
	}
	log.Fatalf("%s: could not find unused field number <= 1000000", m.MessageName)
	return -1
}

func processFile(path string) {
	fd, err := os.Open(path)
	if err != nil {
		log.Fatalf("Open %v: %v", path, err)
	}

	parsed, err := protoparser.Parse(fd,
		protoparser.WithDebug(false),
		protoparser.WithFilename(filepath.Base(path)))
	if err != nil {
		log.Fatalf("Parse %v: %v", path, err)
	}
	pp, err := protoparser.UnorderedInterpret(parsed)
	if err != nil {
		log.Fatalf("Interpret %v: %v", path, err)
	}
	fmt.Printf("Next free IDs\n")
	for _, m := range pp.ProtoBody.Messages {
		fmt.Printf("%s: %v\n", m.MessageName, getNextID(m))
	}
	fd.Close()
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: protofieldnum [foo.proto bar.proto ...]\n")
		os.Exit(1)
	}

	flag.Parse()
	if len(flag.Args()) < 1 {
		flag.Usage()
	}

	for _, path := range flag.Args() {
		processFile(path)
	}
}

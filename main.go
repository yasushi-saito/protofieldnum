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

func getNextID(m *unordered.Message) int {
	used := map[int]bool{}
	for _, f := range m.MessageBody.Fields {
		n, err := strconv.Atoi(f.FieldNumber)
		if err != nil {
			log.Fatalf("%s.%s: Could not parse field number '%s': %v",
				m.MessageName, f.FieldName, f.FieldNumber, err)
		}
		used[n] = true
	}
	for _, oneof := range m.MessageBody.Oneofs {
		for _, f := range oneof.OneofFields {
			n, err := strconv.Atoi(f.FieldNumber)
			if err != nil {
				log.Fatalf("%s.%s: Could not parse field number '%s': %v",
					m.MessageName, f.FieldName, f.FieldNumber, err)
			}
			used[n] = true
		}
	}
	for _, r := range m.MessageBody.Reserves {
		for _, rg := range r.Ranges {
			// log.Printf("Range: %v", rg)
			start, err := strconv.Atoi(rg.Begin)
			if err != nil {
				log.Fatalf("%s %v: Could not parse begin reserve spec: %v",
					m.MessageName, rg, err)
			}
			end := start
			if rg.End != "" {
				if end, err = strconv.Atoi(rg.Begin); err != nil {
					log.Fatalf("%s %v: Could not parse end reserve spec: %v",
						m.MessageName, rg, err)
				}
			}
			if end < start {
				log.Fatalf("%s %v: end <= start", m.MessageName, rg)
			}
			for i := start; i <= end; i++ {
				used[i] = true
			}
		}
	}
	i := 1
	for {
		if i > 1000 {
			log.Fatalf("%s: could not find unused field number <= 1000", m.MessageName)
		}
		if _, ok := used[i]; !ok {
			return i
		}
		i++
	}
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

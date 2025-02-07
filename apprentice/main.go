package main

import (
	"fmt"
	"log"
	"os"

	"github.com/postfix/golibmagic/parser"

	"github.com/postfix/golibmagic/interpreter"

	"github.com/postfix/golibmagic/util"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s TARGET\n", os.Args[0])
		os.Exit(1)
	}

	target := os.Args[1]

	r, err := os.Open(target)
	if err != nil {
		panic(err)
	}

	stats, err := r.Stat()
	if err != nil {
		panic(err)
	}

	sr := util.NewSliceReader(r, 0, stats.Size())

	book := make(parser.Spellbook)
	pc := &parser.ParseContext{
		Logf: log.Printf,
	}
	err = pc.ParseAll("Magdir", book)
	if err != nil {
		panic(err)
	}

	ic := &interpreter.InterpretContext{
		Logf: log.Printf,
		Book: book,
	}

	res, err := ic.Identify(sr)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s: %s\n", target, util.MergeStrings(res))
}

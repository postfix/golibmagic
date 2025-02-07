package golibmagic

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/postfix/golibmagic/interpreter"
	"github.com/postfix/golibmagic/parser"
	"github.com/postfix/golibmagic/util"
)

func doIdentify() error {
	magdir := *identifyArgs.magdir

	NoLogf := func(format string, args ...interface{}) {}

	Logf := func(format string, args ...interface{}) {
		fmt.Println(fmt.Sprintf(format, args...))
	}

	pctx := &parser.ParseContext{
		Logf: NoLogf,
	}

	if *appArgs.debugParser {
		pctx.Logf = Logf
	}

	book := make(parser.Spellbook)
	err := pctx.ParseAll(magdir, book)
	if err != nil {
		return errors.WithStack(err)
	}

	target := *identifyArgs.target
	targetReader, err := os.Open(target)
	if err != nil {
		panic(err)
	}

	defer targetReader.Close()

	stat, _ := targetReader.Stat()

	ictx := &interpreter.InterpretContext{
		Logf: NoLogf,
		Book: book,
	}

	if *appArgs.debugInterpreter {
		ictx.Logf = Logf
	}

	sr := util.NewSliceReader(targetReader, 0, stat.Size())

	result, err := ictx.Identify(sr)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s: %s\n", target, util.MergeStrings(result))

	return nil
}

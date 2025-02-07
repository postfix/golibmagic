package golibmagic

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/postfix/golibmagic/compiler"
	"github.com/postfix/golibmagic/parser"
)

func doCompile() error {
	magdir := *compileArgs.magdir

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

	err = compiler.Compile(book, *compileArgs.output, *compileArgs.chatty, *compileArgs.emitComments, *compileArgs.pkg)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

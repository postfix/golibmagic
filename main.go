package main

import (
	"os"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/go-errors/errors"
	"github.com/itchio/butler/comm"
)

var (
	app = kingpin.New("wizardry", "A magic parser/interpreter/compiler")

	compileCmd  = app.Command("compile", "Compile a set of magic files into one .go file")
	identifyCmd = app.Command("identify", "Use a magic file to identify a target file")
)

var appArgs = struct {
	debugParser      *bool
	debugInterpreter *bool
}{
	app.Flag("debug-parser", "Turn on verbose parser output").Bool(),
	app.Flag("debug-interpreter", "Turn on verbose interpreter output").Bool(),
}

var identifyArgs = struct {
	magdir *string
	target *string
}{
	identifyCmd.Arg("magdir", "the folder of magic files to compile").Required().String(),
	identifyCmd.Arg("target", "path of the the file to identify").Required().String(),
}

var compileArgs = struct {
	magdir *string
	chatty *bool
}{
	compileCmd.Arg("magdir", "the folder of magic files to compile").Required().String(),
	compileCmd.Flag("chatty", "generate prints on every rule match").Bool(),
}

func main() {
	app.HelpFlag.Short('h')
	app.Author("Amos Wenger <amos@itch.io>")

	cmd, err := app.Parse(os.Args[1:])
	if err != nil {
		ctx, _ := app.ParseContext(os.Args[1:])
		app.FatalUsageContext(ctx, "%s\n", err.Error())
	}

	switch kingpin.MustParse(cmd, err) {
	case compileCmd.FullCommand():
		must(doCompile())
	case identifyCmd.FullCommand():
		must(doIdentify())
	}
}

func must(err error) {
	if err != nil {
		switch err := err.(type) {
		case *errors.Error:
			comm.Die(err.ErrorStack())
		default:
			comm.Die(err.Error())
		}
	}
}

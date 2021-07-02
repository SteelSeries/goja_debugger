package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/dop251/goja"
	"github.com/dop251/goja/parser"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"github.com/evanw/esbuild/pkg/api"
)

var dbg *goja.Debugger
var runtime *goja.Runtime

func main() {
	inspect := false
	// Possible values for liveInfo: pc, line, ""
	liveInfo := "pc"
	filename := ""

	if len(os.Args) == 2 {
		filename = os.Args[1]
	} else if len(os.Args) == 3 {
		inspect = (os.Args[1] == "inspect")
		filename = os.Args[2]
	} else if len(os.Args) == 4 {
		inspect = (os.Args[1] == "inspect")
		liveInfo = os.Args[2]
		filename = os.Args[3]
	} else {
		fmt.Printf(cmdHelp, os.Args[0])
		os.Exit(1)
	}

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}

	if inspect {
		// Generate sourceamp on-the-fly
		content = generateSourceMap(filename, string(content))

		fmt.Println("Welcome to Goja debugger")
		fmt.Println("Type 'help' or 'h' for list of commands.")
	}

	printer := console.PrinterFunc(func(s string) {
		prefix := ""
		if inspect {
			prefix = "< "
		}
		fmt.Printf("%s%s\n", prefix, s)
	})

	loader := func(requestedPath string) ([]byte, error) {
		if requestedPath != "" && inspect {
			fmt.Printf("%sLoaded sourcemap from: %s%s\n", GrayColor, requestedPath, ResetColor)
		}
		return nil, nil
	}

	_, err = parser.ParseFile(nil, filename, string(content), 0, parser.WithSourceMapLoader(loader))
	if err != nil {
		log.Fatal(err)
		os.Exit(3)
	}

	runtime = goja.New()

	if inspect {
		dbg = runtime.EnableDebugMode()
	}

	registry := new(require.Registry)
	registry.Enable(runtime)
	registry.RegisterNativeModule("console", console.RequireWithPrinter(printer))
	console.Enable(runtime)

	go func() {
		if inspect {
			reader := bufio.NewReader(os.Stdin)

			reason, resume := dbg.WaitToActivate()
			printDebuggingReason(reason)
			for {
				fmt.Printf("debug%s> ", getInfo(liveInfo))
				userInput, _ := reader.ReadString('\n')
				// convert CRLF to LF
				userInput = strings.Replace(userInput, "\n", "", -1)
				if !repl(userInput) {
					resume()
					reason, resume = dbg.WaitToActivate()
					printDebuggingReason(reason)
				}
			}
		}
	}()

	runtime.RunScript(filename, string(content))
	if err != nil {
		log.Fatal(err)
		os.Exit(4)
	}
}

var cmdHelp = `
%s [inspect] [line|pc] <filename>

inspect: enable debugging
line|pc: show line number or program counter at debug prompt
`[1:]

func printDebuggingReason(reason goja.ActivationReason) {
	if reason == goja.ProgramStartActivation {
		fmt.Printf("Break on start in %s:%d\n", dbg.Filename(), dbg.Line())
	} else if reason == goja.BreakpointActivation {
		fmt.Printf("Break on breakpoint in %s:%d\ns", dbg.Filename(), dbg.Line())
	} else {
		fmt.Printf("Break on debugger statement in %s:%d\n", dbg.Filename(), dbg.Line())
	}
}

func getInfo(liveInfo string) string {
	info := ""
	switch liveInfo {
	case "pc":
		info = fmt.Sprintf("[%d]", dbg.PC())
	case "line":
		info = fmt.Sprintf("[%d]", dbg.Line())
	default:
		info = fmt.Sprintf("[%d]", dbg.PC())
	}
	return info
}

func generateSourceMap(filename string, src string) []byte {
	result := api.Transform(src, api.TransformOptions{
		Sourcemap:  api.SourceMapInline,
		Sourcefile: filename,
	})

	if len(result.Errors) > 0 {
		fmt.Println(result.Errors)
	}

	return result.Code
}

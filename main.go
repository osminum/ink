package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
)

const VERSION = "0.1.0"

const HELP_MSG = `
Ink is a minimal, functional programming language.
	ink v%s

By default, ink interprets from stdin.
	ink < main.ink
Run an ink script on files with -input.
	ink -input main.ink
Run from the command line with -eval.
	ink -eval "f := () => out('hi'), f()"

`

// for input files flag parsing
type inkFiles []string

func (i *inkFiles) Set(val string) error {
	*i = append(*i, val)
	return nil
}

func (i *inkFiles) String() string {
	return strings.Join(*i, ", ")
}

func main() {
	flag.Usage = func() {
		fmt.Printf(HELP_MSG, VERSION)
		flag.PrintDefaults()
	}

	// cli arguments
	verbose := flag.Bool("verbose", false, "Log all interpreter debug information")
	debugLexer := flag.Bool("debug-lex", false, "Log lexer output")
	debugParser := flag.Bool("debug-parse", false, "Log parser output")
	dump := flag.Bool("dump", false, "Dump global frame after eval")

	version := flag.Bool("version", false, "Print version string and exit")
	help := flag.Bool("help", false, "Print help message and exit")

	repl := flag.Bool("repl", false, "Run as an interactive repl")
	eval := flag.String("eval", "", "Evaluate argument as an Ink script")

	var files inkFiles
	flag.Var(&files, "input", "Source code to execute, can be invoked multiple times")

	flag.Parse()

	// if asked for version, disregard everything else
	if *version {
		fmt.Println("ink", VERSION)
		os.Exit(0)
	} else if *help {
		flag.Usage()
		os.Exit(0)
	}

	// execution context
	ctx := Context{}
	ctx.Init()

	// abstraction for executing ink code from a file at a given path
	execFile := func(path string) error {
		// read from file
		input, errors := ctx.ExecStream(
			*debugLexer || *verbose,
			*debugParser || *verbose,
			*dump || *verbose,
		)

		file, err := os.Open(path)
		defer file.Close()
		if err != nil {
			logSafeErr(
				ErrSystem,
				fmt.Sprintf("could not open %s for execution:\n\t-> %s", path, err),
			)
			close(input)
			return err
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			for _, char := range scanner.Text() {
				input <- char
			}
			input <- '\n'
		}
		close(input)

		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			for range ctx.ValueStream {
				// continue
			}
			wg.Done()
		}()
		go func() {
			for e := range errors {
				logSafeErr(e.reason, fmt.Sprintf("in %s\n\t-> ", path)+e.message)
				wg.Done()
				return
			}
			wg.Done()
		}()

		wg.Wait()
		return nil
	}

	if *repl {
		// run interactively in a repl
		reader := bufio.NewReader(os.Stdin)

		shouldExit := false
		for !shouldExit {
			// green arrow
			fmt.Printf(ANSI_GREEN_BOLD + "> " + ANSI_RESET)
			text, err := reader.ReadString('\n')

			if err != nil {
				logErrf(ErrSystem, "unexpected stop to input:\n\t-> %s", err.Error())
			}

			switch {
			// specialized introspection / observability directives
			//	in repl session
			case strings.HasPrefix(text, "@dump"):
				ctx.Dump()
			case strings.HasPrefix(text, "@clear"):
				fmt.Printf("[2J[H")
			case strings.HasPrefix(text, "@exit"):
				shouldExit = true
			case strings.HasPrefix(text, "@load "):
				path := strings.Trim(text[5:], " \t\n")
				err := execFile(path)
				if err == nil {
					logDebugf("loaded file:\n\t-> %s", path)
				}

			default:
				input, errors := ctx.ExecStream(
					*debugLexer || *verbose,
					*debugParser || *verbose,
					*dump || *verbose,
				)

				for _, char := range text {
					input <- char
				}
				close(input)

				wg := sync.WaitGroup{}
				wg.Add(2)
				go func() {
					for v := range ctx.ValueStream {
						logInteractive(v.String())
					}
					wg.Done()
				}()
				go func() {
					for e := range errors {
						logSafeErr(e.reason, e.message)
						wg.Done()
						return
					}
					wg.Done()
				}()

				wg.Wait()
			}
		}

		os.Exit(0)
	} else if *eval != "" {
		input, errors := ctx.ExecStream(
			*debugLexer || *verbose,
			*debugParser || *verbose,
			*dump || *verbose,
		)

		for _, char := range *eval {
			input <- char
		}
		close(input)

		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			for range ctx.ValueStream {
				// continue
			}
			wg.Done()
		}()
		go func() {
			for e := range errors {
				logErr(e.reason, e.message)
				wg.Done()
				return
			}
			wg.Done()
		}()

		wg.Wait()
	} else if len(files) > 0 {
		// read from file
		for _, path := range files {
			execFile(path)
		}
	} else {
		// read from stdin
		input, errors := ctx.ExecStream(
			*debugLexer || *verbose,
			*debugParser || *verbose,
			*dump || *verbose,
		)

		inputReader := bufio.NewReader(os.Stdin)
		for {
			char, _, err := inputReader.ReadRune()
			if err != nil {
				break
			}
			input <- char
		}
		close(input)

		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			for range ctx.ValueStream {
				// continue
			}
			wg.Done()
		}()
		go func() {
			for e := range errors {
				logErr(e.reason, e.message)
				wg.Done()
				return
			}
			wg.Done()
		}()

		wg.Wait()
	}
}

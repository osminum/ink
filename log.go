package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	ANSI_RESET       = "[0;0m"
	ANSI_BLUE        = "[34;22m"
	ANSI_GREEN       = "[32;22m"
	ANSI_YELLOW      = "[33;22m"
	ANSI_RED         = "[31;22m"
	ANSI_BLUE_BOLD   = "[34;1m"
	ANSI_GREEN_BOLD  = "[32;1m"
	ANSI_YELLOW_BOLD = "[33;1m"
	ANSI_RED_BOLD    = "[31;1m"
)

func logDebug(args ...string) {
	fmt.Println(ANSI_BLUE_BOLD + "debug: " + ANSI_BLUE + strings.Join(args, " ") + ANSI_RESET)
}

func logDebugf(s string, args ...interface{}) {
	logDebug(fmt.Sprintf(s, args...))
}

func logInteractive(args ...string) {
	fmt.Println(ANSI_GREEN + strings.Join(args, " ") + ANSI_RESET)
}

func logInteractivef(s string, args ...interface{}) {
	logInteractive(fmt.Sprintf(s, args...))
}

func logWarn(args ...string) {
	fmt.Println(ANSI_YELLOW_BOLD + "warn: " + ANSI_YELLOW + strings.Join(args, " ") + ANSI_RESET)
}

func logWarnf(s string, args ...interface{}) {
	logWarn(fmt.Sprintf(s, args...))
}

func logSafeErr(reason int, args ...string) {
	errStr := "error"
	switch reason {
	case ErrSyntax:
		errStr = "syntax error"
	case ErrRuntime:
		errStr = "runtime error"
	case ErrSystem:
		errStr = "system error"
	case ErrAssert:
		errStr = "invariant violation"
	default:
		errStr = "error"
	}
	fmt.Println(ANSI_RED_BOLD + errStr + ": " + ANSI_RED + strings.Join(args, " ") + ANSI_RESET)
}

func logErr(reason int, args ...string) {
	logSafeErr(reason, args...)
	os.Exit(reason)
}

func logErrf(reason int, s string, args ...interface{}) {
	logErr(reason, fmt.Sprintf(s, args...))
}
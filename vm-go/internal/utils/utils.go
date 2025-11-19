package utils

import (
	"fmt"
	"os"
	"time"
)

const (
	Red    = "\033[0;31m"
	Green  = "\033[0;32m"
	Blue   = "\033[0;34m"
	Gray   = "\033[0;90m"
	Yellow = "\033[0;33m"
	NC     = "\033[0m"
)

func Log(format string, args ...interface{}) {
	fmt.Printf("[%s] %s\n", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
}

func Error(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%sERROR:%s %s\n", Red, NC, fmt.Sprintf(format, args...))
	os.Exit(1)
}

func CheckRoot() {
	if os.Geteuid() == 0 {
		fmt.Fprintf(os.Stderr, "%s[ERROR]%s This script should NOT be run as root or with sudo.\n", Red, NC)
		os.Exit(1)
	}
}

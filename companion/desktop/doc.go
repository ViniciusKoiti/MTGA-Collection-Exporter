//go:build !windows

// The desktop shell is Windows-only: it needs the WebView2 runtime,
// and building the Wails frontend on the Linux CI runners would
// demand GTK/WebKit system libraries the pipeline does not carry.
// Every functional file is tagged windows; this stub keeps ./...
// builds green on the other platforms.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "desktop: the companion shell runs on Windows only")
	os.Exit(1)
}

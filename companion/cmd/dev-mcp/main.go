// Command dev-mcp is the DEVELOPMENT-ONLY stdio MCP server (OpenSpec
// add-graph-workflow-harness, task 7.1). It ships in no product
// package — an arch test scans every packaging manifest to keep it
// that way — and refuses to start against any target not classified
// as development or staging (production and ambiguity are denied
// before any connection).
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/devmcp"
)

func main() {
	host := flag.String("host", "", "target host (development/staging only)")
	dsn := flag.String("dsn", "", "target database DSN")
	marker := flag.String("environment", "development",
		"explicit environment marker")
	flag.Parse()
	target := devmcp.Alvo{Host: *host, DSN: *dsn, Marcador: *marker}
	if err := devmcp.AutorizaConexao(target); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := devmcp.ServeStdio(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

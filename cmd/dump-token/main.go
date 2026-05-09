// Outil interne : imprime le token Freebox app sur stdout pour scripting.
// Usage : go run ./cmd/dump-token
// Securite : ne jamais committer la sortie ; ne jamais lancer en presence d'un audit log non-controle.
package main

import (
	"fmt"
	"os"

	"github.com/mipsou/mcp-freebox/internal/wincred"
)

func main() {
	t, err := wincred.Read("freebox-mcp")
	if err != nil {
		fmt.Fprintf(os.Stderr, "dump-token: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(t)
}

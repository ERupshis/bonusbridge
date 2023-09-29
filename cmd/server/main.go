package main

import (
	"fmt"
	"os"

	"github.com/erupshis/bonusbridge/internal/logger"
)

func main() {
	log, err := logger.CreateZapLogger("info")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create logger: %v", err)
	}

}

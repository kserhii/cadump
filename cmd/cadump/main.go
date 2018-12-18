package main

import (
	"fmt"
	"time"

	"cadump/cadump"
)

func main() {
	start := time.Now()
	cadump.ProcessScan()
	fmt.Printf("Done in %s\n", time.Since(start))
}

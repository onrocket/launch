package main

import (
	"fmt"

	"github.com/onrocket/launch/binfiles"
	"github.com/onrocket/launch/config"
)

func main() {
	fmt.Printf("Move along please, nothing to see here ...\n")
	config.LoadConfig()
	binfiles.LoadBinFiles()
}

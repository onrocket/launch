package main

import (
	"fmt"

	"github.com/onrocket/launch/binfiles"
	"github.com/onrocket/launch/config"
)

func main() {
	config.LoadConfig()
	binfiles.LoadBinFiles()
}

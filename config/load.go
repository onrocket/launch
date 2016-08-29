package config

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

func printError(err error) {
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("==> Error: %s\n", err.Error()))
	}
}

func printOutput(outs []byte) {
	if len(outs) > 0 {
		fmt.Printf("==> Output: %s", string(outs))
		switch {
		case strings.Contains(string(outs), "text"):
			fmt.Println("TXT")
		case strings.Contains(string(outs), "executable"):
			fmt.Println("EXE")
		}
		fmt.Printf("\n")
	}
}

// LoadConfig is currently called by main but will be moved back to this module
func LoadConfig() {

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(usr.HomeDir)

	// TODO : create and populate onrocket launch directory if not already there
	//        and optionaly accept an alternative directory if specified as a
	//        command line parameter
	searchDir := usr.HomeDir + "/.onrocket/launch"

	fmt.Println("+++++=STILL definitely nothing to see here, move along ...")

	fileList := []string{}
	err = filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return nil
	})
	if err != nil {
		fmt.Printf("bumber : %s\n", err)
		os.Exit(1)
	}

	for _, file := range fileList {
		choppedFile := strings.Replace(file, searchDir, "", -1)
		fmt.Printf(">>>>>[%s][%s]\n", file, choppedFile)

		// Create an *exec.Cmd
		cmd := exec.Command("file", "-b", file)

		// Combine stdout and stderr
		//printCommand(cmd)
		output, err := cmd.CombinedOutput()
		printError(err)
		printOutput(output)

	}

}

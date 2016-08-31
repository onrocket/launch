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

func printOutput(outs []byte) (ftype string) {
	if len(outs) > 0 {
		switch {
		case strings.Contains(string(outs), "text"):
			ftype = "TXT"
		case strings.Contains(string(outs), "executable"):
			ftype = "EXT"
		}
	}
	return
}

// LoadConfig is currently called by main but will be moved back to this module
func LoadConfig() {

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(usr.HomeDir)

	// TODO : create and populate onrocket launch directory if not already there
	//        and optionaly accept an alternative directory if specified as a
	//        command line parameter
	searchDir := usr.HomeDir + "/.onrocket/launch"

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

		// Create an *exec.Cmd
		cmd := exec.Command("file", "-b", file)

		// Combine stdout and stderr
		output, err := cmd.CombinedOutput()
		printError(err)
		ftype := printOutput(output)
		if ftype == "TXT" {
			fmt.Printf("got ftype [%s] back from printOutput\n", ftype)
			fmt.Printf("\t     file[%s]\n\tetcd path[%s]\n\n", file, choppedFile)
		}

	}

}

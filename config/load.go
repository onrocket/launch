package config

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

func exitIfError(err error, reason string) {
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("==> Error: %s [%s]\n", err.Error(), reason))
		os.Exit(1)
	}
}

func parseFileCmdOutput(outs []byte) (fType string) {
	if len(outs) > 0 {
		switch {
		case strings.Contains(string(outs), "text"):
			fType = "TXT"
		case strings.Contains(string(outs), "executable"):
			fType = "EXT"
		}
	}
	return
}

func LaunchConfigPath() string {

	usr, err := user.Current()
	exitIfError(err, "trying to get current user")
	// TODO : create and populate onrocket launch directory if not already there
	//        and optionaly accept an alternative directory if specified as a
	//        command line parameter
	searchDir := usr.HomeDir + "/.onrocket/Launch"
	return searchDir

}

func launchConfigFiles(searchDir string) []string {
	fileList := []string{}
	err := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return nil
	})
	exitIfError(err, fmt.Sprintf("can't read directory [%s]", searchDir))
	return fileList
}

func findConfigs(fileList []string, searchDir string) {
	for _, file := range fileList {
		fileCmd := exec.Command("file", "-b", file)
		fileCmdOutput, err := fileCmd.CombinedOutput()
		exitIfError(err, "running file command through exec")
		fType := parseFileCmdOutput(fileCmdOutput)
		if fType == "TXT" {
			uploadFileOrConfig(file, searchDir)
		}
	}
}

func uploadFileOrConfig(file, searchDir string) {
	choppedFile := strings.Replace(file, searchDir, "", -1)
	ext := filepath.Ext(file)
	fmt.Printf("\t     file[%s]\n", file)
	fmt.Printf("\tetcd path[%s]\n", choppedFile)
	fmt.Printf("\t      ext[%s]\n\n", ext)
	if (ext == ".csv") || ext == ".CSV" {
		parseUploadCSVData(file, searchDir, choppedFile)
	} else {
		uploadFileData(file, choppedFile)
	}
}

func parseUploadCSVData(file, searchDir, choppedFile string) {
	baseKeyAddr := strings.Replace(choppedFile, ".csv", "", -1)
	fmt.Printf("about to upload CSV [%s]\n\n", baseKeyAddr)
	csvText, err := ioutil.ReadFile(file)
	exitIfError(err, "trying to open file")
	r := csv.NewReader(strings.NewReader(string(csvText)))
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			exitIfError(err, "csv parsing error")
		}
		etcdPath := fmt.Sprintf("%s/%s", baseKeyAddr, record[0])
		fmt.Printf("record :: [%s]=>[%s]\n", etcdPath, record[1])
		uploadToEtcd(etcdPath, []byte(record[1]))
	}
}

func uploadFileData(file, choppedFile string) {
	fmt.Printf("about to upload text file complete\n\n")
	plainText, err := ioutil.ReadFile(file)
	exitIfError(err, "trying to open file")
	uploadToEtcd(choppedFile, plainText)
}

func uploadToEtcd(key string, val []byte) {
	//TODO :: establish settings for location of etcd - not just hard coded
	//        to localhost
	cfg := client.Config{
		Endpoints: []string{"http://127.0.0.1:2379"},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target
		// endpoint is unavailable - TODO : make this configurable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	exitIfError(err, "configuring etcd")
	kapi := client.NewKeysAPI(c)
	_, err = kapi.Set(context.Background(), key, string(val), nil)
	exitIfError(err, fmt.Sprintf("trying to upload key [%s] to val [%s] etcd", key, val))
}

// LoadConfig is currently called by main but will be moved back to this module
func LoadConfig() {
	searchDir := LaunchConfigPath()
	fileList := launchConfigFiles(searchDir)
	findConfigs(fileList, searchDir)
}

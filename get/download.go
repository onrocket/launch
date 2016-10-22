package get

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/nutrun/lentil"
	"github.com/onrocket/launch/config"
	"golang.org/x/net/context"
)

var (
	bean     *lentil.Beanstalkd
	hostName *string
)

type Location struct {
	x, y int
	name string
}

type Template struct {
	ARG01, ARG02, ARG03, ARG04, ARG05,
	ARG06, ARG07, ARG08, ARG09, ARG10,
	ARG11, ARG12, ARG13, ARG14, ARG15,
	ARG16, ARG17, ARG18, ARG19, ARG20 string
}

type OnRocket struct {
	template Template
	mom      map[string]map[string]string
}

// JSONJobStr used to marshal incoming JSON
type JSONJobStr struct {
	ID   string `json:"id"`
	User string `json:"user"`
	Job  string `json:"job"`
}

func init() {
	var err error
	bean, err = lentil.Dial("0.0.0.0:11300")
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println("Hostname reported by kernel : ", hostName)
}

// DownloadConfig - called by main to get a local copy of config from etcd
func (rkt *OnRocket) DownloadConfig(nodeName, serviceName string) {

	fmt.Printf("\n> nodeName[%s] serviceName[%s]\n\n", nodeName, serviceName)

	rkt.mom = map[string]map[string]string{}

	fullPath := "/DC1/Config/host/" + nodeName + "/" + serviceName + "/template-variables"
	rkt.downloadConfigFromPath(fullPath, "template_variables")

	fullPath = "/DC1/Config/host/" + nodeName + "/" + serviceName + "/variables"
	rkt.downloadConfigFromPath(fullPath, "variables")

	fullPath = "/DC1/Sequence/Hosts/" + nodeName
	rkt.downloadConfigFromPath(fullPath, "script_tags")

	fullPath = "/DC1/Sequence/Scripts"
	rkt.downloadConfigFromPath(fullPath, "scripts")

	fullPath = "/DC1/Sequence/Templates"
	rkt.downloadConfigFromPath(fullPath, "templates")

	rkt.buildScriptsFromTemplates(nodeName, serviceName)

}

func (rkt *OnRocket) buildScriptsFromTemplates(nodeName, serviceName string) {

	for j := range rkt.mom["template_variables"] {
		val := rkt.mom["template_variables"][j]

		switch j {
		case "ARG01":
			rkt.template.ARG01 = val
		case "ARG02":
			rkt.template.ARG02 = val
		case "ARG03":
			rkt.template.ARG03 = val
		case "ARG04":
			rkt.template.ARG04 = val
		case "ARG05":
			rkt.template.ARG05 = val
		case "ARG06":
			rkt.template.ARG06 = val
		case "ARG07":
			rkt.template.ARG07 = val
		case "ARG08":
			rkt.template.ARG08 = val
		case "ARG09":
			rkt.template.ARG09 = val
		case "ARG10":
			rkt.template.ARG10 = val
		case "ARG11":
			rkt.template.ARG11 = val
		case "ARG12":
			rkt.template.ARG12 = val
		case "ARG13":
			rkt.template.ARG13 = val
		case "ARG14":
			rkt.template.ARG14 = val
		case "ARG15":
			rkt.template.ARG15 = val
		case "ARG16":
			rkt.template.ARG16 = val
		case "ARG17":
			rkt.template.ARG17 = val
		case "ARG18":
			rkt.template.ARG18 = val
		case "ARG19":
			rkt.template.ARG19 = val
		case "ARG20":
			rkt.template.ARG20 = val
		}
	}

	for j := range rkt.mom["templates"] {
		fullPath := "/DC1/Sequence/Templates/" + j
		templateShortName := "template_" + j

		rkt.downloadConfigFromPath(fullPath, templateShortName)

		t, err := template.New("template_name").Parse(rkt.mom[templateShortName][j])
		if err != nil {
			log.Fatal("bigo problemo ", err)
		}
		buf := new(bytes.Buffer)
		t.Execute(buf, rkt.template)
		scriptName := j
		scriptName = strings.Replace(scriptName, ".tmpl", "", -1)

		scriptContents := fmt.Sprintf("%s", buf)

		scriptDir := scriptDirectory(nodeName, serviceName)

		if _, err := os.Stat(scriptDir); os.IsNotExist(err) {
			fmt.Printf("creating [%s]\n", scriptDir)
			err = os.MkdirAll(scriptDir, 0777)
			if err != nil {
				log.Fatal(err)
			}
		}

		filePath := scriptDir + "/" + scriptName
		f, err := os.Create(filePath)
		if err != nil {
			log.Fatal(err)
		}
		n3, err := f.WriteString(scriptContents)
		if err != nil {
			log.Fatal(err)
		}
		err = os.Chmod(filePath, 0755)
		hashStr := GetMD5Hash(filePath)

		fmt.Printf("wrote %d bytes to %s [%s][%s][%s][%s]\n", n3, filePath, j, nodeName, serviceName, hashStr)

	}

}

func scriptDirectory(nodeName, serviceName string) (scriptDir string) {
	searchDir := config.LaunchConfigPath()
	scriptDir = searchDir + "/DC1/Config/host/" + nodeName + "/" + serviceName + "/scripts"
	return scriptDir
}

func (rkt *OnRocket) downloadConfigFromPath(fullPath string, category string) {
	//set confuration structure with address of etcd server etc.
	cfg := client.Config{
		Endpoints:               []string{"http://127.0.0.1:2379"},
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}
	// connect to etcd service and query path passed through fullPath parameter
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi := client.NewKeysAPI(c)
	getopt := client.GetOptions{
		Recursive: true,
	}
	resp, err := kapi.Get(context.Background(), fullPath, &getopt)
	if err != nil {
		log.Fatal(err)
	}
	// create a map to store records
	rec := map[string]string{}
	// iterate over the sub values of the path passed through fullPath parameter
	for _, node := range resp.Node.Nodes {
		_, keyval := path.Split(node.Key)
		rec[string(keyval)] = string(node.Value)
	}
	//store record data into a map of maps with key passed by category parameter
	rkt.mom[string(category)] = rec
}

func (rkt *OnRocket) LentilListener() {

	for {
		_, err := bean.Watch("JobRequests")
		if err != nil {
			log.Fatal(err)
		}

		job, e := bean.Reserve()
		if e != nil {
			log.Fatal(e)
		}

		log.Printf("JOB ID: %d, JOB BODY: %s", job.Id, job.Body)

		runJobRequest(job.Body)

		e = bean.Delete(job.Id)
		if e != nil {
			log.Fatal(e)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func runJobRequest(body []byte) {

	var hostName string
	envHost, isSet := os.LookupEnv("HOST")
	if isSet {
		hostName = envHost
	} else {
		var err error
		hostName, err = os.Hostname()
		if err != nil {
			log.Fatal(err)
		}
	}

	res := JSONJobStr{}
	json.Unmarshal(body, &res)
	fmt.Printf("     user : %s\n", res.User)
	fmt.Printf("       id : %s\n", res.ID)
	fmt.Printf("      job : %s\n", res.Job)
	fmt.Printf(" hostname : %s\n", hostName)

	scriptDir := scriptDirectory(hostName, res.ID)
	fmt.Printf("scriptDir:[%s]\n", scriptDir)

}

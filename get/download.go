package get

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/nutrun/lentil"
	"github.com/onrocket/launch/config"
	"golang.org/x/net/context"
	"gopkg.in/mgo.v2"
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

// JSONJobStr used to marshal job JSON
type JSONJobStr struct {
	ID   string `json:"id"`
	User string `json:"user"`
	Job  string `json:"job"`
}

type JSONLogStr struct {
	ID   string `json:"id"`
	User string `json:"user"`
	Job  string `json:"job"`
	Log  string `json:"log"`
}

type JobLog struct {
	EpochTime int    `json:"epochtime"`
	User      string `json:"user"`
	JobName   string `json:"jobname"`
	JobStatus string `json:"jobstatus"`
}

func init() {
	var err error
	bean, err = lentil.Dial("0.0.0.0:11300")
	if err != nil {
		log.Fatalf("failed to connected to lentil :: %s", err)
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
				log.Fatalf("failed to create directory :: %s", err)
			}
		}

		filePath := scriptDir + "/" + scriptName
		f, err := os.Create(filePath)
		if err != nil {
			log.Fatalf("failed to create dir :: %s\n", err)
		}
		n3, err := f.WriteString(scriptContents)
		if err != nil {
			log.Fatalf("failed to write string :: %s\n", err)
		}
		err = os.Chmod(filePath, 0755)
		if err != nil {
			log.Fatalf("failed to make executable :: %s\n", err)
		}
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
		log.Fatalf("failed to connect to etcd :: %s \n", err)
	}
	kapi := client.NewKeysAPI(c)
	getopt := client.GetOptions{
		Recursive: true,
	}
	resp, err := kapi.Get(context.Background(), fullPath, &getopt)
	if err != nil {
		log.Fatalf("failed to get fullPath [%s] :: %s\n", fullPath, err)
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

func (rkt *OnRocket) LentilLogger(secint int, userID, jobName, jobText string) {

	mylog := JobLog{
		EpochTime: secint,
		User:      userID,
		JobName:   jobName,
		JobStatus: jobText,
	}
	l, err := json.Marshal(mylog)
	if err != nil {
		log.Fatalf("json marshal failed with [%s]\n", err)
	}
	ls := string(l)

	lentilBean, err := lentil.Dial("0.0.0.0:11300")
	if err != nil {
		log.Fatalf("failed to connected to lentil :: %s", err)
	}
	defer lentilBean.Quit()

	lentilBean.Use("JobLog")

	jobId, err := lentilBean.Put(0, 0, 60, []byte(ls))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("INSERTED JOB ID: %d\n", jobId)

	log.Printf("BOOM [%d][%s][%s][%s]\n>>>%s\n\n", secint, userID, jobName, jobText, ls)
}

func (rkt *OnRocket) LentilLogListener() {
	fmt.Printf("Log Listenter fired\n")
	for {
		_, err := bean.Watch("JobLog")
		if err != nil {
			log.Fatal(err)
		}

		job, e := bean.Reserve()
		if e != nil {
			log.Fatal(e)
		}

		log.Printf("JOB ID: %d, JOB BODY: %s", job.Id, job.Body)
		res := JobLog{}
		e = json.Unmarshal(job.Body, &res)
		if e != nil {
			log.Fatalf("error unmarshalling : %s", e)
		}

		e = bean.Delete(job.Id)

		if e != nil {
			log.Fatalf("error deleteing from beanstalk : %s", e)
		}

		// LOG TO FILE
		fmt.Printf("     epoch : %d\n", res.EpochTime)
		fmt.Printf("      user : %s\n", res.User)
		fmt.Printf("       job : %s\n", res.JobName)
		fmt.Printf("    output : %s\n", res.JobStatus)

		logFile := "/tmp/launch.log"
		_, err = os.Stat(logFile)
		if os.IsNotExist(err) {
			file, err := os.Create(logFile)
			if err != nil {
				fmt.Printf("Error creating file %s\n", err)
			}
			file.Close()
		}

		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}

		defer f.Close()
		dateStr := fmt.Sprintf("%s", time.Unix(int64(res.EpochTime), 0))
		text := fmt.Sprintf("%s :: %d :: %s %s :: %s\n", dateStr, res.EpochTime, res.User, res.JobName, res.JobStatus)
		if _, err = f.WriteString(text); err != nil {
			panic(err)
		}
		writeToPipeCmd(string(job.Body))

		time.Sleep(100 * time.Millisecond)
	}
}

func writeToPipeCmd(jsonText string) {

	pathToCmd := os.Getenv("STDIN_CMD")
	fmt.Printf("we have [%s] for envCommand [STDIN_CMD]\n", pathToCmd)
	//pathToCmd := "/home/jon/dev/onrocket/bin/stdin"

	if _, err := os.Stat(pathToCmd); os.IsNotExist(err) {
		fmt.Printf("file [%s] is not found\n", pathToCmd)
	} else {
		grepCmd := exec.Command(pathToCmd)

		grepIn, err := grepCmd.StdinPipe()
		if err != nil {
			log.Fatalf("error opening pipe to stdin [%s]\n", err)
		}
		grepOut, err := grepCmd.StdoutPipe()
		if err != nil {
			log.Fatalf("error opening pipe to stdout [%s]\n", err)
		}
		grepCmd.Start()
		grepIn.Write([]byte(jsonText))
		grepIn.Close()
		grepBytes, _ := ioutil.ReadAll(grepOut)
		grepCmd.Wait()
		fmt.Printf("here is the ouput of our pipe from stdout : [%s]\n", string(grepBytes))
	}
}

func (rkt *OnRocket) LentilJobListener() {

	for {
		_, err := bean.Watch("JobRequests")
		if err != nil {
			log.Fatal(err)
		}

		job, e := bean.Reserve()
		if e != nil {
			log.Fatal(e)
		}

		log.Printf("INCOMING JOB ID: %d, JOB BODY: %s", job.Id, job.Body)

		rkt.runJobRequest(job.Body)

		e = bean.Delete(job.Id)
		if e != nil {
			log.Fatalf("error deleteing from beanstalk : %s", e)
		}
		fmt.Println(">>>> deleted a beanstalkd queue entry")
		time.Sleep(100 * time.Millisecond)
	}
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func printCommand(cmd *exec.Cmd) {
	fmt.Printf("==> Executing: %s\n", strings.Join(cmd.Args, " "))
}

func printError(err error) {
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("==> Error: %s\n", err.Error()))
	}
}

func printOutput(outs []byte) {
	if len(outs) > 0 {
		fmt.Printf("==> Output: %s\n", string(outs))
	}
}

func (rkt *OnRocket) runJobRequest(body []byte) {

	var hostName string
	envHost, isSet := os.LookupEnv("HOST")
	if isSet {
		hostName = envHost
	} else {
		var err error
		hostName, err = os.Hostname()
		if err != nil {
			log.Fatalf("failed to get hostame :: %s", err)
		}
	}

	res := JSONJobStr{}
	json.Unmarshal(body, &res)
	fmt.Printf("     user : %s\n", res.User)
	fmt.Printf("       id : %s\n", res.ID)
	fmt.Printf("      job : %s\n", res.Job)
	fmt.Printf(" hostname : %s\n", hostName)

	scriptDir := scriptDirectory(hostName, res.ID)
	scriptToRun := scriptDir + "/" + res.Job
	fmt.Printf("scriptDir:[%s]\n", scriptDir)
	fmt.Printf("to run:[%s]\n", scriptToRun)

	cmdArgs := []string{""}

	cmd := exec.Command(scriptToRun, cmdArgs...)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(cmdReader)
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	defer session.Close()
	go func(userID string) {
		for scanner.Scan() {
			fmt.Printf("stdout > [%s]\n", scanner.Text())

			now := time.Now()
			secs := now.Unix()
			secint := int(secs)

			// Optional. Switch the session to a monotonic behavior.
			//session.SetMode(mgo.Monotonic, true)
			//c := session.DB("myapp").C("joblog")
			logText := fmt.Sprintf("%s", scanner.Text())
			checkSubStr := "[METEOR]"
			if strings.Contains(logText, checkSubStr) {
				newLogText := strings.Replace(logText, "[METEOR]", "", -1)
				//newLogText := logText
				//err = c.Insert(&JobLog{EpochTime: secint, User: userID, JobName: "siteCopy", JobStatus: newLogText})
				//if err != nil {
				//	log.Fatal(err)
				//}
				rkt.LentilLogger(secint, userID, "siteCopy", newLogText)
				fmt.Printf("[%d][%s], %s [%s]\n", secint, userID, "siteCopy", newLogText)
			}
		}
		fmt.Printf("\n\nDone. [%s]\n", userID)
	}(res.ID)

	err = cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error starting Cmd", err)
		os.Exit(1)
	}

	err = cmd.Wait()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error waiting for Cmd", err)
		os.Exit(1)
	}
	fmt.Println("at end of running job request")

}

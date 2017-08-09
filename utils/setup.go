package utils

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	"github.com/fatih/color"
)

var (
	srv *http.Server
)

func StartServer() *http.Server {
	srv = &http.Server{Addr: ":8000"}
	http.Handle("/", http.FileServer(http.Dir("./package")))

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
			log.Printf("Httpserver: ListenAndServe() error: %s", err)
		}

		color.Blue("%sHTTP file server listening at: 0.0.0.0:8000", CheckSymbol)
	}()

	return srv
}

func RunSetup(script string, ch chan int, args ...string) {
	//cmd := exec.Command("bash", "-s", script)
	cmd := exec.Command(script, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	defer stdout.Close()
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	if strings.Contains(script, "master") {
		go saveLog(stdout, true)
	} else {
		go saveLog(stdout, false)
	}

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				log.Printf("Exit Status: %d", status.ExitStatus())
			}
		} else {
			log.Fatalf("cmd.Wait: %v", err)
		}
		ch <- 1
	} else {
		ch <- 0
	}
}

func matchToken(buf []byte) {
	if strings.Contains(string(buf), "kubeadm join --token") {
		re := regexp.MustCompile("kubeadm join --token [0-9a-z.]*")
		result := re.Find(buf)

		//Get the token string
		token := strings.Split(string(result), " ")[3]
		ioutil.WriteFile("./.k8s-token", []byte(token), os.ModeAppend)
		color.Green("%sMaster token %s saved into .k8s.token file.", CheckSymbol, token)
	}
}

func outputProgress(buf []byte) {
	re := regexp.MustCompile("KUBEKIT_OUTPUT .*.")
	results := re.FindAll(buf, -1)

	if results != nil {
		for _, r := range results {
			//Get the token string
			output := strings.Replace(string(r), "KUBEKIT_OUTPUT ", "", 1)
			color.Blue(output)
		}
	}
}

func saveLog(stdout io.ReadCloser, saveToken bool) {

	fd, _ := os.OpenFile("install.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)

	defer fd.Close()

	for {
		var n int
		buf := make([]byte, 1024)
		n, err := stdout.Read(buf)

		if err != nil {
			//End of output
			break
		}

		//Output the install progress
		if strings.Contains(string(buf), "KUBEKIT_OUTPUT") {
			outputProgress(buf)
		}

		if saveToken {
			matchToken(buf)
		}

		fd.WriteString(string(buf[:n]))
	}
}

func SetupDocker() bool {
	color.Blue("Start to install docker engine...\r\n\r\n")
	ch := make(chan int)

	go RunSetup("./package/docker.sh", ch)
	if <-ch == 1 {
		color.Red("\r\n\r\n%sFailed to install docker engine...\r\n\r\n", CrossSymbol)
		return false
	}

	color.Green("\r\n\r\n%sDocker engine installed...\r\n\r\n", CheckSymbol)
	return true
}

func SetupMaster() bool {
	color.Blue("Start to initialize Kubernetes master node...\r\n\r\n")
	ch := make(chan int)

	go RunSetup("./package/master.sh", ch, "master")
	if <-ch == 1 {
		color.Red("\r\n\r\n%sFailed to initialize Kubernetes master node...\r\n\r\n", CrossSymbol)
		return false
	}

	color.Green("\r\n\r\n%sKubernetes master node initialized...\r\n\r\n", CheckSymbol)
	return true
}

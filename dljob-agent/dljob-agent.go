package main

import (
	"bytes"
	"dlengine/utils/conv"
	"dlengine/utils/file"
	"dljob/dljob-agent/config"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func getIP() string {
	cmd := exec.Command("/bin/bash", "-c", "ifconfig eth0 | grep 'inet addr:' | grep -v '127.0.0.1' | cut -d: -f2 | awk '{print $1}'")
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput
	cmd.Run()

	ip := string(cmdOutput.Bytes())
	ip = strings.Trim(ip, "\n")
	return ip
}

func syncResult(resultPath string) {
	//rsync -R results/delong/result.txt.10.77.96.56 10.77.96.56::dljob
	args := "rsync -R " + resultPath + " " + config.CENTER
	cmd := exec.Command("/bin/bash", "-c", args)
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	var waitStatus syscall.WaitStatus
	err := cmd.Run()
	if err != nil {
		fmt.Println(error.Error(err))
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus = exitError.Sys().(syscall.WaitStatus)
			fmt.Printf("Exit code: %d\n", waitStatus.ExitStatus())
		}
		//Left result for checking
		return
	}
	os.Remove(resultPath)
}

func execute(filePath string, fileContent multipart.File) {
	f, err := os.Create(filePath)
	if err != nil {
		fmt.Println(error.Error(err))
		return
	}
	defer f.Close()
	io.Copy(f, fileContent)

	cmd := exec.Command("/bin/bash", filePath)
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput
	cmd.Run()

	//Result file
	ip := getIP()
	scriptName := file.Basename(filePath, true)
	prefix := config.RESULTPATH + scriptName
	resultPath := prefix + "/result.txt." + ip

	err = os.MkdirAll(prefix, 0777)
	if err != nil {
		fmt.Println(error.Error(err))
		os.Remove(filePath)
		return
	}

	r, err := os.Create(resultPath)
	if err != nil {
		fmt.Println(error.Error(err))
		os.Remove(filePath)
		return
	}
	defer r.Close()
	r.Write(cmdOutput.Bytes())

	syncResult(resultPath)

	//Clean up
	os.Remove(filePath)
}

func initRoutes() {
	http.HandleFunc("/dljob", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		//Parse
		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			fmt.Println(error.Error(err))
			return
		}

		var fileName string
		var filePath string

		if r.MultipartForm != nil {
			fileName = r.MultipartForm.Value["filename"][0]
		}
		if fileName != "" {
			filePath = config.FILEPATH + fileName
			if file.IsExisit(filePath) {
				fmt.Fprintf(w, "Script [%s] is processing now\n", fileName)
				return
			} else {
				w.Write([]byte("OK"))
			}
		} else {
			w.Write([]byte("Script file is empty"))
			return
		}

		fileContent, _, err := r.FormFile("script")
		if err != nil {
			fmt.Println(error.Error(err))
			return
		}
		defer fileContent.Close()

		go execute(filePath, fileContent)
	})
}

func start(port int) {
	localHost := "0.0.0.0"
	addr := localHost + ":" + conv.GetString(port)

	s := &http.Server{
		Addr:           addr,
		MaxHeaderBytes: 1 << 30,
	}

	log.Fatalln(s.ListenAndServe())
}

func main() {
	port := flag.Int("p", 0, "listen port")
	version := flag.Bool("v", false, "show version")

	flag.Parse()

	if *version {
		fmt.Println(config.VERSION)
		os.Exit(0)
	}

	if *port == 0 {
		fmt.Println(config.NOPORT)
		os.Exit(0)
	}

	initRoutes()

	go start(*port)

	select {}
}

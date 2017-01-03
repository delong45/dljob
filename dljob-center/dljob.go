package main

import (
	"dlengine/dlhttp"
	"dlengine/utils/file"
	"dljob/dljob-center/config"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func getDomainIPS(domain string) string {
	switch strings.ToUpper(domain) {
	case "MT":
		return config.MT
	case "API":
		return config.API
	case "API2":
		return config.API2
	case "PCT":
		return config.PCT
	case "TEST":
		return config.TEST
	default:
		log.Fatalln("No such domain")
	}

	return ""
}

func singleOrder(ip string, scriptFile string) error {
	name := file.Basename(scriptFile, false)
	params := map[string]string{
		"filename": name,
	}

	addr := ip + ":" + config.AGENT_PORT
	uri := "http://" + addr + "/dljob"
	fmt.Printf("%s\n", uri)

	request, err := dlhttp.FileUploadRequest(uri, params, "script", scriptFile)
	if err != nil {
		fmt.Printf("FileUploadRequest Error:%s\n", error.Error(err))
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		fmt.Printf("Client request Error:%s\n", error.Error(err))
		return err
	}

	if resp.StatusCode == http.StatusOK {
		var bodyContent []byte
		defer resp.Body.Close()
		bodyContent, _ = ioutil.ReadAll(resp.Body)

		fmt.Printf("Send order to %s: %s\n", ip, string(bodyContent))
	} else {
		fmt.Printf("Order %s excute %s failed: %d\n", ip, scriptFile, resp.StatusCode)
	}

	return nil
}

func order(domain string, scriptFile string) {
	if !file.IsExisit(scriptFile) {
		log.Fatalln("No such script file")
	}

	ips := getDomainIPS(domain)
	ipArray := strings.Split(ips, " ")
	for _, ip := range ipArray {
		err := singleOrder(ip, scriptFile)
		if err != nil {
			fmt.Printf("Order %s excute %s failed\n", ip, scriptFile)
		}
	}
}

func main() {
	domain := flag.String("d", "", "specify domain")
	scriptFile := flag.String("f", "", "specify script file path")
	version := flag.Bool("v", false, "show version")

	flag.Parse()

	if *version {
		fmt.Println(config.VERSION)
		os.Exit(0)
	}

	if *domain == "" {
		fmt.Println(config.NODOMAIN)
		os.Exit(0)
	}

	if *scriptFile == "" {
		fmt.Println(config.NOSCRIPT)
		os.Exit(0)
	}

	order(*domain, *scriptFile)
}

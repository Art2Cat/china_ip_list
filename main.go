package main

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"

	"math"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const apnicURL = "http://ftp.apnic.net/apnic/stats/apnic/delegated-apnic-latest"
const apincIPListFile = "./apinc_ip_list.txt"
const ipipURL = "https://raw.githubusercontent.com/17mon/china_ip_list/master/china_ip_list.txt"
const ipipIPListFile = "./ip_list.txt"
const outPutFile = "./china_ip_list.txt"

func main() {
	taskJob()
}

func init() {

	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	// log.SetOutput(os.Stdout)
	var file, err = os.OpenFile("./error.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Could Not Open Log File : " + err.Error())
	}
	log.SetOutput(file)
	// Only log the warning severity or above.
	log.SetLevel(log.WarnLevel)
}

func taskJob() {
	initJob()
	apincIPList := parseChinaIPFromApinc()
	ipipList := openIpipFile()
	finalIPList := mergeSliceWithOutDuplicate(ipipList, apincIPList)
	file, err := os.OpenFile(outPutFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer file.Close()
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	dataWriter := bufio.NewWriter(file)

	for _, data := range finalIPList {
		_, _ = dataWriter.WriteString(data + "\n")
	}

	dataWriter.Flush()

	cmd := exec.Command("git", "add", "china_ip_list.txt")
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	cmd = exec.Command("git", "commit", "-m", "'update china_ip_list.txt'")
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	cmd = exec.Command("git", "push")
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func initJob() {
	err := os.Chdir("./china_ip_list")
	if err != nil {
		log.Fatal(err)
	}
	os.Remove(outPutFile)
	os.Remove(ipipIPListFile)
	os.Remove(apincIPListFile)
	downloadFile(ipipIPListFile, ipipURL)
	downloadFile(apincIPListFile, apnicURL)
}

func mergeSliceWithOutDuplicate(a []string, b []string) []string {

	check := make(map[string]int)
	d := append(a, b...)
	res := make([]string, 0)
	for _, val := range d {
		check[val] = 1
	}

	for letter := range check {
		res = append(res, letter)
	}

	return res
}

func openIpipFile() []string {
	f, err := os.Open(ipipIPListFile)
	if err != nil {
		log.Fatal(err)
	}
	r4 := bufio.NewReader(f)
	ipList := make([]string, 0)
	for {
		line, _, err := r4.ReadLine()

		if err == io.EOF {
			break
		}
		s := string(line)
		ipList = append(ipList, s)
	}
	return ipList
}

func parseChinaIPFromApinc() []string {
	f, err := os.Open(apincIPListFile)
	if err != nil {
		log.Fatal(err)
	}
	r4 := bufio.NewReader(f)
	ipList := make([]string, 0)
	for {
		line, _, err := r4.ReadLine()

		if err == io.EOF {
			break
		}

		s := string(line)
		if strings.Contains(s, "apinc") || strings.Contains(s, "|CN|ipv4|") {
			split := strings.Split(s, `|`)
			i, _ := strconv.Atoi(split[4])
			mask := 32 - math.Log(float64(i)/math.Log(2))
			ip := split[3] + "/" + strconv.Itoa(int(mask))
			ipList = append(ipList, ip)
		}
	}
	return ipList
}

func downloadFile(filepath string, url string) (err error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Panic(err)
	}
	defer func() {
		e := resp.Body.Close()
		if e != nil {
			err = e
		}
	}()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		log.Panic(err)
	}
	defer func() {
		e := out.Close()
		if e != nil {
			err = e
		}
	}()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Panic(err)
	}
	return
}

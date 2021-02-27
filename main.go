package main

import (
	"bufio"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const ApnicUrl = "http://ftp.apnic.net/apnic/stats/apnic/delegated-apnic-latest"
const ApincIpListFile = "apinc_ip_list.txt"
const IpipUrl = "https://raw.githubusercontent.com/17mon/china_ip_list/master/china_ip_list.txt"
const IpipIpListFile = "ip_list.txt"
const OutPutFile = "china_ip_list.txt"

func main() {
	apincIpList := parseChinaIpFromApinc()
	ipipList := openIpipFile()
	finalIpList := mergeSliceWithOutDuplicate(ipipList, apincIpList)
	file, err := os.OpenFile(OutPutFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer file.Close()
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	dataWriter := bufio.NewWriter(file)

	for _, data := range finalIpList {
		_, _ = dataWriter.WriteString(data + "\n")
	}

	dataWriter.Flush()
	cmd := exec.Command("git", "add", "china_ip_list.txt")
	err = cmd.Run()
	if err != nil {
		log.Println(err)
	}

	cmd = exec.Command("git", "commit", "'update china_ip_list.txt'")
	err = cmd.Run()
	if err != nil {
		log.Println(err)
	}
	cmd = exec.Command("git", "push")
	err = cmd.Run()
	if err != nil {
		log.Println(err)
	}
}

func init() {
	os.Remove(OutPutFile)
	os.Remove(IpipIpListFile)
	os.Remove(ApincIpListFile)
	downloadFile(IpipIpListFile, IpipUrl)
	downloadFile(ApincIpListFile, ApnicUrl)
}

func mergeSliceWithOutDuplicate(a []string, b []string) []string {

	check := make(map[string]int)
	d := append(a, b...)
	res := make([]string, 0)
	for _, val := range d {
		check[val] = 1
	}

	for letter, _ := range check {
		res = append(res, letter)
	}

	return res
}

func openIpipFile() []string {
	f, err := os.Open(IpipIpListFile)
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

func parseChinaIpFromApinc() []string {
	f, err := os.Open(ApincIpListFile)
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

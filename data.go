package main

import (
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
)

type Metrics struct {
	Name string
}

type LinkPair struct {
	Pod1      string
	Pod2      string
	Bandwidth string
	Laterncy  string
}

const (
	vNICSymbol = "10.233"
	LinkSymbol = "link"
	Test       = "eth0"
	SrcSymbol  = "src"
)

func (m Metrics) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	name := "LinkStatus"
	linkTypes := []string{}
	linkTypes = append(linkTypes, "networkName", "star1", "star2", "bandwidth", "laterncy", "BFR")
	metrics := GetTopology()

	for _, metric := range metrics {
		data := []string{}
		data = append(data, m.Name, metric.Pod1, metric.Pod2, metric.Bandwidth, metric.Laterncy, "0.1%")
		if s, err := GeneratePromData(name, linkTypes, data); err == nil {
			fmt.Fprint(w, s)
		}
	}

}

func GetNICTraffic(NIC string) string {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("./data.sh %s", NIC))
	stdout, _ := cmd.CombinedOutput()
	traffic := string(stdout)
	print(traffic)
	return traffic
}

func GetLinkLaterncy(target string) string {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("ping -c 5 -i 0.2 %s", target))
	stdout, _ := cmd.CombinedOutput()
	outStr := string(stdout)
	print(outStr)
	if strings.Contains(outStr, "100% packet loss") {
		return "--"
	}
	outLines := strings.Split(outStr, "\n")
	var laterncy string
	for _, line := range outLines {
		if strings.Contains(line, "rtt") {
			words := strings.Split(line, "/")
			laterncy = words[4]
		}
	}
	return laterncy
}

func GetTopology() []LinkPair {
	cmd := exec.Command("sh", "-c", "echo $PODNAME")
	stdout, _ := cmd.CombinedOutput()
	podName := string(stdout)
	print(podName)
	cmd = exec.Command("sh", "-c", "ip route")
	stdout, _ = cmd.CombinedOutput()
	outStr := string(stdout)
	print(outStr)
	outLines := strings.Split(outStr, "\n")
	linkPairs := make([]LinkPair, 0)
	starIPs := make(map[string]string)
	for _, line := range outLines {
		if len(line) != 0 {
			if strings.Contains(line, vNICSymbol) {
				words := strings.Split(line, " ")
				starIPs[words[2]] = words[0]
			} else if strings.Contains(line, LinkSymbol) && strings.Contains(line, SrcSymbol) {
				words := strings.Split(line, " ")
				if value, ok := starIPs[words[2]]; ok {
					linkPairs = append(linkPairs, LinkPair{Pod1: podName, Pod2: words[2], Bandwidth: GetNICTraffic(words[2]), Laterncy: GetLinkLaterncy(value)})
				}
			}
		}
	}
	for key, value := range starIPs {
		print(key)
		print(value)
	}
	return linkPairs
}

//	func (m Metrics) ServeHTTP(w http.ResponseWriter, req *http.Request) {
//		name := "StarStatus"
//		types := []string{}
//		types = append(types, "name", "timestamp", "CPU", "memory", "disk", "status")
//		data := []string{}
//		data = append(data, "node1", fmt.Sprint(time.Now().Unix()), "0.01%", "100MB", "200MB", "up")
//		if s, err := GeneratePromData(name, types, data); err == nil {
//			fmt.Fprint(w, s)
//		}
//	}
func main() {
	metric := Metrics{Name: "Network1"}
	http.Handle("/metrics", metric)
	http.ListenAndServe(":2112", nil)
}

func GeneratePromData(name string, types []string, datas []string) (string, error) {
	var result string
	if len(types) != len(datas) {
		return result, errors.New("lens of types is difference from lens of datas for prometheus type")
	}
	result = fmt.Sprintf("# HELP %s data structure\n# TYPE %s counter\n", name, name)
	result += name + "{"
	for index := range types {
		result += fmt.Sprintf("%s=\"%s\",", types[index], datas[index])
	}
	result = result[:len(result)-1] + "} 1"
	return result, nil
}

// CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o link_exporter data.go

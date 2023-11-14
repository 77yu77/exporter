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
	Pod1 string
	Pod2 string
}

const (
	vNICName   = "sdn"
	LinkSymbol = "link"
)

func (m Metrics) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	name := "LinkStatus"
	linkTypes := []string{}
	linkTypes = append(linkTypes, "networkName", "star1", "star2", "bandwidth", "laterncy", "BFR")
	metrics := GetTopology()

	for _, metric := range metrics {
		data := []string{}
		data = append(data, m.Name, metric.Pod1, metric.Pod2, "100MB/sec", "5ms", "0.1%")
		if s, err := GeneratePromData(name, linkTypes, data); err == nil {
			fmt.Fprint(w, s)
		}
	}

}

func GetNICTraffic(NIC string) string {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("./data.sh %s", NIC))
	stdout, _ := cmd.CombinedOutput()
	traffic := string(stdout)
	return traffic
}

func GetLinkLaterncy(target string) string {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("ping %s", target))
	stdout, _ := cmd.CombinedOutput()
	laterncy := string(stdout)
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
	for _, line := range outLines {
		if len(line) != 0 {
			if strings.Contains(line, LinkSymbol) {
				if strings.Contains(line, vNICName) {
					words := strings.Split(line, " ")
					linkPairs = append(linkPairs, LinkPair{Pod1: podName, Pod2: words[2]})
				}
			}
		}
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

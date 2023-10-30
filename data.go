package main

import (
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
)

type Metrics struct {
	Pod1 string
	Pod2 string
}

const (
	vNICName        = "sdn"
	GolbalIPSegment = "10.233"
	LinkSymbol      = "link"
)

func (m Metrics) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	name := "LinkStatus"
	types := []string{}
	types = append(types, "star1", "star2", "bandwidth", "laterncy", "BFR")
	metrics := GetTopology()

	for _, metric := range metrics {
		data := []string{}
		data = append(data, metric.Pod1, metric.Pod2, "100MB/sec", "5ms", "0.1%")
		if s, err := GeneratePromData(name, types, data); err == nil {
			fmt.Fprint(w, s)
		}
	}
}

func GetNICTraffic(NIC string) string {
	cmd := exec.Command("./data.sh", NIC)
	stdout, _ := cmd.CombinedOutput()
	traffic := string(stdout)
	return traffic
}

func GetLinkLaterncy(target string) string {
	cmd := exec.Command("ping", target)
	stdout, _ := cmd.CombinedOutput()
	laterncy := string(stdout)
	return laterncy
}

func GetTopology() []Metrics {
	cmd := exec.Command("echo", "$PODNAME")
	stdout, _ := cmd.CombinedOutput()
	podName := string(stdout)
	cmd = exec.Command("ip", "route")
	stdout, _ = cmd.CombinedOutput()
	outStr := string(stdout)
	outLines := strings.Split(outStr, "\n")
	metrics := make([]Metrics, 0)
	var metric Metrics
	// for _, word := range words {
	// 	if strings.Contains(word, vNICName) {
	// 		if metric.Pod1 == "" {
	// 			metric.Pod1 = word
	// 		} else {
	// 			metric.Pod2 = word
	// 		}
	// 	}
	// }
	NICs := make([]string, 0)
	for _, line := range outLines {
		if strings.Contains(line, GolbalIPSegment) {
			words := strings.Split(line, " ")
			NIC := words[len(words)-1]
			NICs = append(NICs, NIC)

		} else if strings.Contains(line, LinkSymbol) {
			for _, NIC := range NICs {
				if strings.Contains(line, NIC) {
					metrics = append(metrics, Metrics{Pod1: podName, Pod2: NIC})
					break
				}
			}
		}
		metrics = append(metrics, metric)
	}
	return metrics
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
	metric := Metrics{}
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

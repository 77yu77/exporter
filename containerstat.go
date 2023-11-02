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

type StarStatus struct {
	Name   string
	CPU    string
	Memory string
	Mempct string
	Disk   string
}

const (
	ContainerSymbol = "satellite"
)

func (m Metrics) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	name := "StarStatus"
	ContainerTypes := []string{}
	ContainerTypes = append(ContainerTypes, "networkName", "starName", "CPU", "memory", "mempct", "disk")
	stars := GetContainerMessage()

	for _, star := range stars {
		data := []string{}
		data = append(data, m.Name, star.Name, star.CPU, star.Memory, star.Mempct, star.Disk)
		if s, err := GeneratePromData(name, ContainerTypes, data); err == nil {
			fmt.Fprint(w, s)
		}
	}

}

func GetContainerMessage() []StarStatus {
	diskMessages := GetDiskMessage()
	cmd := exec.Command("docker", "stats", "--no-stream", "|", "grep", ContainerSymbol)
	stdout, _ := cmd.CombinedOutput()
	output := string(stdout)
	containerMessages := strings.Split(output, "\n")
	starStatuses := make([]StarStatus, 0)

	for _, line := range containerMessages {
		datas := strings.Split(line, " ")
		var starStatus StarStatus
		starStatus.Name = datas[1]
		starStatus.Disk = diskMessages[datas[0]][4]
		starStatus.CPU = datas[2]
		starStatus.Memory = datas[3]
		starStatus.Mempct = datas[6]
		starStatuses = append(starStatuses, starStatus)
	}

	return starStatuses
}

// get all the symbolized container's disk message
func GetDiskMessage() map[string][]string {
	cmd := exec.Command("docker", "system", "df", "-v", "|", "grep", ContainerSymbol)
	stdout, _ := cmd.CombinedOutput()
	diskMessage := strings.Split(string(stdout), "\n")
	diskMessages := make(map[string][]string, 0)
	for _, line := range diskMessage {
		words := strings.Split(line, " ")
		diskMessages[words[0]] = words
	}
	return diskMessages
}

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

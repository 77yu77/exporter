package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type Metrics struct {
	Name    string
	ImageId string
}

type StarStatus struct {
	Name   string
	CPU    string
	Memory string
	Mempct string
}

func (m Metrics) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	name := "StarStatus"
	ContainerTypes := []string{}
	ContainerTypes = append(ContainerTypes, "networkName", "Name", "CPU", "memory", "mempct")
	stars := GetContainerMessage(m.ImageId)

	for _, star := range stars {
		data := []string{}
		data = append(data, m.Name, star.Name, star.CPU, star.Memory, star.Mempct)
		print(data)
		print("\n")
		if s, err := GeneratePromData(name, ContainerTypes, data); err == nil {
			fmt.Fprint(w, s)
		}
	}

}

func GetContainerMessage(ImageId string) []StarStatus {
	print("container\n")
	// diskMessages := GetDiskMessage(ImageId)
	cmd := exec.Command("bash", "-c", fmt.Sprintf("docker ps | grep %s", ImageId))
	stdout, _ := cmd.CombinedOutput()
	output := string(stdout)
	// print(output)
	containerMessages := strings.Split(output, "\n")
	containerIds := make(map[string]bool)
	for _, message := range containerMessages {
		if len(message) != 0 {
			words := strings.Split(message, " ")
			containerIds[words[0]] = true
		}
	}

	cmd = exec.Command("bash", "-c", fmt.Sprintf("docker stats --no-stream"))
	stdout, _ = cmd.CombinedOutput()
	output = string(stdout)
	// print(output)
	containerMessages = strings.Split(output, "\n")

	starStatuses := make([]StarStatus, 0)

	for _, line := range containerMessages {
		datas := strings.Fields(line)
		if len(datas) != 0 {
			if _, ok := containerIds[datas[0]]; ok {
				var starStatus StarStatus
				words := strings.Split(datas[1], "_")
				starStatus.Name = words[1]
				starStatus.CPU = datas[2]
				starStatus.Memory = datas[3]
				starStatus.Mempct = datas[6]
				// fmt.Println(starStatus)
				starStatuses = append(starStatuses, starStatus)
			}

		}
	}

	return starStatuses
}

// get all the symbolized container's disk message
func GetDiskMessage(ImageId string) map[string][]string {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("docker system df -v | grep %s", ImageId))
	stdout, _ := cmd.CombinedOutput()
	outStr := string(stdout)
	diskMessage := strings.Split(outStr, "\n")
	diskMessages := make(map[string][]string, 0)
	for _, line := range diskMessage {
		words := strings.Fields(line)
		if len(words) != 0 {
			diskMessages[words[0]] = words
		}
	}
	return diskMessages
}

func main() {
	args := os.Args
	metric := Metrics{args[1], args[2]}
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

// c2be9bc309db

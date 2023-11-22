package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	args := os.Args
	Bandwidth := args[1]
	cmd := exec.Command("sh", "-c", fmt.Sprintf("echo $SERVERIP"))
	stdout, _ := cmd.CombinedOutput()
	output := string(stdout)
	ServerIP := strings.Fields(output)[0]
	print(ServerIP)
	cmd = exec.Command("sh", "-c", fmt.Sprintf("ps"))
	stdout, _ = cmd.CombinedOutput()
	output = string(stdout)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if len(line) != 0 && strings.Contains(line, "iperf3 -c") {
			words := strings.Fields(line)
			for _, word := range words {
				println(word)
			}
			cmd := exec.Command("sh", "-c", fmt.Sprintf("kill -9 %s", words[0]))
			stdout, _ := cmd.CombinedOutput()
			output = string(stdout)
			time.Sleep(time.Duration(1) * time.Second)
			break
		}
	}
	cmd = exec.Command("sh", "-c", fmt.Sprintf("iperf3 -c %s -b %s -p 5202 –i 1 -t 1000 &", ServerIP, Bandwidth))
	cmd.Run()
}

// CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o client client.go

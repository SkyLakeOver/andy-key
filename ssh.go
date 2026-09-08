package main

import (
	"fmt"
	"os/exec"
	"strconv"
)

func ExecuteSSHCommand(username, password, ip string, port int, command string) (string, error) {
	cmd := exec.Command("sshpass", "-p", password,
		"ssh", "-o", "StrictHostKeyChecking=no", "-o", "ConnectTimeout=15",
		"-p", strconv.Itoa(port),
		fmt.Sprintf("%s@%s", username, ip),
		command)

	output, err := cmd.CombinedOutput()
	result := string(output)
	if err != nil {
		result += fmt.Sprintf("\n[ERROR: %v]", err)
	}
	return result, nil
}
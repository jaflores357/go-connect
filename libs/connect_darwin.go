
// +build darwin

package connect

import (
	"strings"
	"os/exec"
)

func wrapInQuotes(text string) string {
	return "\"" + text + "\""
}

func SshConn(conn string) {
	

	ssh_command := "ssh -o UserKnownHostsFile=/dev/null -o ServerAliveInterval=30 -o StrictHostKeyChecking=no "+conn
	exec_cmd := "create tab with default profile command " + wrapInQuotes(ssh_command)
	application := wrapInQuotes("iTerm2")

	script := []string{"tell application", application, "\n","tell current window","\n", exec_cmd, "\n", "end tell", "\n", "end tell"}
	
	command := strings.Join(script, " ")
	out := exec.Command("osascript", "-e", command)
	output, err := out.CombinedOutput()

	_ = output 
	_ = err
	//fmt.Println(output)
	//fmt.Println(err)
	
}
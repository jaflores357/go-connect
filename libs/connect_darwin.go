
// +build darwin

package connect

import (
	"strings"
	"os/exec"
	"os"
	"fmt"
)

func Username() string {
	return os.Getenv("USER")
}

func wrapInQuotes(text string) string {
	return "\"" + text + "\""
}

func SshConn(conn string, username string, sshkey string) {
	
	if sshkey != "" {
		sshkey = " -i " + sshkey
	}
	
	ssh_command := "ssh -o UserKnownHostsFile=/dev/null -o ServerAliveInterval=30 -o StrictHostKeyChecking=no "+username+"@"+conn+sshkey
	terminal_app := os.Getenv("TERM_PROGRAM")
	command := ""

	if terminal_app == "iTerm.app" {
	
		exec_cmd := "create tab with default profile command " + wrapInQuotes(ssh_command)
		application := wrapInQuotes("iTerm2")
		script := []string{"tell application", application, "\n","tell current window","\n", exec_cmd, "\n", "end tell", "\n", "end tell"}
		command = strings.Join(script, " ")

	} else if terminal_app == "Apple_Terminal" {

		exec_cmd := "do script " + wrapInQuotes(ssh_command)
		application := wrapInQuotes("Terminal")
		script := []string{"tell application ", application, "\n", "tell application \"System Events\" to keystroke \"t\" using {command down}", "\n", exec_cmd, " in front window", "\n","end tell"}
		command = strings.Join(script, " ")
		
	} else {

		fmt.Println("Teminal not supported (yet): "+terminal_app)
		os.Exit(0)		

	}


	out := exec.Command("osascript", "-e", command)
	output, err := out.CombinedOutput()

	_ = output 
	_ = err
	
}
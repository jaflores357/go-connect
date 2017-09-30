
// +build linux

package connect

import (
	"fmt"
	"os"
)

func Username() string {
	return os.Getenv("USER")
}

func wrapInQuotes(text string) string {
	return "\"" + text + "\""
}

func SshConn(conn string) {

	ssh_command := "ssh -o UserKnownHostsFile=/dev/null -o ServerAliveInterval=30 -o StrictHostKeyChecking=no "+conn
	command := "gnome-terminal -x bash -c " + wrapInQuotes(ssh_command)
	
	_ = exec.Command(command)
	
}
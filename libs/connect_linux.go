
// +build linux

package connect

import (
	"os"
	"os/exec"
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
	command := "gnome-terminal -x bash -c " + wrapInQuotes(ssh_command)
	
	_, err := exec.Command("bash", "-c", command).Output()
	if err != nil {
		panic(err)
	}
	
}
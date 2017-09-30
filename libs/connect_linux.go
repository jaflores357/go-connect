
// +build linux

package connect

import (
	"fmt"
	"os"
)

func Username() string {
	return os.Getenv("USER")
}

func SshConn(conn string) {

	_ = exec.Command("xdotool", "key", "ctrl+shift+t")
	_ = exec.Command("echo", conn)

	fmt.Println("ssh linux: ", conn)
	
}

// +build linux

package connect

import "fmt"

func Username() string {
	return os.Getenv("USER")
}

func SshConn(conn string) {

	fmt.Println("ssh linux: ", conn)
	
}
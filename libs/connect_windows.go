
// +build !linux,!darwin !cgo

package connect

import "fmt"

func SshConn(conn string) {

	fmt.Println("ssh windows: ", conn)
	
}
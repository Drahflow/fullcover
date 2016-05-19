// Copyright 2016 by Drahflow. Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

package sender

import (
	"net"
	"fmt"
)

var con net.Conn

func initConnection(receiver string) {
	if con != nil {
		return
	}

	var err error
	con, err = net.Dial("tcp", receiver)
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(con, "POST /coverage HTTP/1.1\r\nHost: localhost\r\nTransfer-Encoding: chunked\r\n\r\n")
}

func ReportFile(receiver string, filename string, source string) {
	initConnection(receiver)

	chunk := fmt.Sprintf("F%d:%s%d:%s", len(filename), filename, len(source), source)
	fmt.Fprintf(con, "%x\r\n%s\r\n", len(chunk), chunk)
}

func ReportBlock(receiver string, filename string, startLine int, startCol int, endLine int, endCol int, numStmt int) {
	initConnection(receiver)

	chunk := fmt.Sprintf("B%d:%s%d:%d:%d:%d:%d:", len(filename), filename, startLine, startCol, endLine, endCol, numStmt)
	fmt.Fprintf(con, "%x\r\n%s\r\n", len(chunk), chunk)
}
func ReportCover(receiver string, filename string, startLine int, startCol int, endLine int, endCol int, numStmt int) {
	initConnection(receiver)

	chunk := fmt.Sprintf("C%d:%s%d:%d:%d:%d:%d:", len(filename), filename, startLine, startCol, endLine, endCol, numStmt)
	fmt.Fprintf(con, "%x\r\n%s\r\n", len(chunk), chunk)
}

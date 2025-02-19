// Copyright 2018 Venil Noronha. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"sync/atomic"
	"time"
)

var _connCount = int64(0)
var _connRecvCount = int64(0)

// main serves as the program entry point
func main() {
	// obtain the port and prefix via program arguments
	port := fmt.Sprintf(":%s", os.Args[1])
	prefix := os.Args[2]

	// create a tcp listener on the given port
	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("failed to create listener, err:", err)
		os.Exit(1)
	}
	fmt.Printf("listening on %s, prefix: %s\n", listener.Addr(), prefix)

	// print stats
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			connCount := atomic.LoadInt64(&_connCount)
			connRecvCount := atomic.SwapInt64(&_connRecvCount, 0)

			if connCount > 0 {
				fmt.Printf("connCount: %d, connRecvCount: %d\n", connCount, connRecvCount)
			}
		}
	}()

	// listen for new connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("failed to accept connection, err:", err)
			continue
		}

		// pass an accepted connection to a handler goroutine
		go handleConnection(conn, prefix)
	}
}

// handleConnection handles the lifetime of a connection
func handleConnection(conn net.Conn, prefix string) {
	atomic.AddInt64(&_connCount, 1)
	defer atomic.AddInt64(&_connCount, -1)

	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		// read client request data
		bytes, err := reader.ReadBytes(byte('\n'))
		if err != nil {
			if err != io.EOF {
				fmt.Println("failed to read data, err:", err)
			}
			return
		}
		atomic.AddInt64(&_connRecvCount, 1)

		// prepend prefix and send as response
		line := fmt.Sprintf("%s %s", prefix, bytes)
		conn.Write([]byte(line))
	}
}

//  Copyright (c) 2020 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package e2e

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

// TCP or UDP connection request
type connectionRequest struct {
	conn net.Conn
	err  error
}

func simpleTCPServer(ctx context.Context, ms *Microservice, addr string, expReqMsg, respMsg string, done chan<- error) {
	// move to the network namespace where server should listen
	exitNetNs := ms.enterNetNs()
	defer exitNetNs()

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		done <- err
		return
	}
	defer listener.Close()

	// accept single connection
	newConn := make(chan connectionRequest, 1)
	go func() {
		conn, err := listener.Accept()
		newConn <- connectionRequest{conn: conn, err: err}
		close(newConn)
	}()

	// wait for connection
	var cr connectionRequest
	select {
	case <-ctx.Done():
		done <- fmt.Errorf("tcp server listening on %s was canceled", addr)
		return
	case cr = <-newConn:
		if cr.err != nil {
			done <- fmt.Errorf("accept failed with: %v", cr.err)
			return
		}
		defer cr.conn.Close()
	}

	// communicate with the client
	commRv := make(chan error, 1)
	go func() {
		defer close(commRv)
		// receive message from the client
		message, err := bufio.NewReader(cr.conn).ReadString('\n')
		if err != nil {
			commRv <- fmt.Errorf("failed to read data from client: %v", err)
			return
		}
		// send response to the client
		_, err = cr.conn.Write([]byte(respMsg + "\n"))
		if err != nil {
			commRv <- fmt.Errorf("failed to send data to client: %v", err)
			return
		}
		// check if the exchanged data are as expected
		message = strings.TrimRight(message, "\n")
		if message != expReqMsg {
			commRv <- fmt.Errorf("unexpected message received from client ('%s' vs. '%s')",
				message, expReqMsg)
			return
		}
		commRv <- nil
	}()

	// wait for the message exchange to execute
	select {
	case <-ctx.Done():
		done <- fmt.Errorf("tcp server listening on %s was canceled", addr)
		return
	case err = <-commRv:
		done <- err
	}

	// do not close until client confirms reception of the message
	<-ctx.Done()
}

func simpleUDPServer(ctx context.Context, ms *Microservice, addr string, expReqMsg, respMsg string, done chan<- error) {
	const maxBufferSize = 1024
	// move to the network namespace where server should listen
	exitNetNs := ms.enterNetNs()
	defer exitNetNs()

	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		done <- err
		return
	}
	defer conn.Close()

	// communicate with the client
	commRv := make(chan error, 1)
	go func() {
		defer close(commRv)
		// receive message from the client
		buffer := make([]byte, maxBufferSize)
		n, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			commRv <- fmt.Errorf("failed to read data from client: %v", err)
			return
		}
		message := string(buffer[:n])
		// send response to the client
		_, err = conn.WriteTo([]byte(respMsg+"\n"), addr)
		if err != nil {
			commRv <- fmt.Errorf("failed to send data to client: %v", err)
			return
		}
		// check if the exchanged data are as expected
		message = strings.TrimRight(message, "\n")
		if message != expReqMsg {
			commRv <- fmt.Errorf("unexpected message received from client ('%s' vs. '%s')",
				message, expReqMsg)
			return
		}
		commRv <- nil
	}()

	// wait for the message exchange to execute
	select {
	case <-ctx.Done():
		done <- fmt.Errorf("udp server listening on %s was canceled", addr)
		return
	case err = <-commRv:
		done <- err
	}

	// do not close until client confirms reception of the message
	<-ctx.Done()
}

func simpleTCPClient(ms *Microservice, addr string, reqMsg, expRespMsg string, timeout time.Duration, done chan<- error) {
	// try to connect with the server
	newConn := make(chan connectionRequest, 1)

	go func() {
		// move to the network namespace from which the connection should be initiated
		exitNetNs := ms.enterNetNs()
		defer exitNetNs()
		start := time.Now()
		for {
			conn, err := net.Dial("tcp", addr)
			if err != nil && time.Since(start) < timeout {
				time.Sleep(checkPollingInterval)
				continue
			}
			newConn <- connectionRequest{conn: conn, err: err}
			break
		}
		close(newConn)
	}()

	simpleTCPOrUDPClient(newConn, addr, reqMsg, expRespMsg, timeout, done)
}

func simpleUDPClient(ms *Microservice, addr string, reqMsg, expRespMsg string, timeout time.Duration, done chan<- error) {
	// try to connect with the server
	newConn := make(chan connectionRequest, 1)

	go func() {
		// move to the network namespace from which the connection should be initiated
		exitNetNs := ms.enterNetNs()
		defer exitNetNs()
		udpAddr, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			newConn <- connectionRequest{conn: nil, err: err}
		} else {
			start := time.Now()
			for {
				conn, err := net.DialUDP("udp", nil, udpAddr)
				if err != nil && time.Since(start) < timeout {
					time.Sleep(checkPollingInterval)
					continue
				}
				newConn <- connectionRequest{conn: conn, err: err}
				break
			}
		}
		close(newConn)
	}()

	simpleTCPOrUDPClient(newConn, addr, reqMsg, expRespMsg, timeout, done)
}

func simpleTCPOrUDPClient(newConn chan connectionRequest, addr, reqMsg, expRespMsg string,
	timeout time.Duration, done chan<- error) {

	// wait for connection
	var cr connectionRequest
	select {
	case <-time.After(timeout):
		done <- fmt.Errorf("connection to %s timed out", addr)
		return
	case cr = <-newConn:
		if cr.err != nil {
			done <- fmt.Errorf("dial failed with: %v", cr.err)
			return
		}
		defer cr.conn.Close()
	}

	// communicate with the server
	commRv := make(chan error, 1)
	go func() {
		defer close(commRv)

		// send message to the server
		_, err := cr.conn.Write([]byte(reqMsg + "\n"))
		if err != nil {
			commRv <- fmt.Errorf("failed to send data to the server: %v", err)
			return
		}
		// listen for reply
		start := time.Now()
		var message string
		for {
			message, err = bufio.NewReader(cr.conn).ReadString('\n')
			if err != nil && time.Since(start) < timeout {
				time.Sleep(checkPollingInterval)
				continue
			}
			if err != nil {
				commRv <- fmt.Errorf("failed to read data from server: %v", err)
				return
			}
			break
		}
		// check if the exchanged data are as expected
		message = strings.TrimRight(message, "\n")
		if message != expRespMsg {
			commRv <- fmt.Errorf("unexpected message received from server ('%s' vs. '%s')",
				message, expRespMsg)
			return
		}

		commRv <- nil
	}()

	// wait for the message exchange to execute
	select {
	case <-time.After(timeout):
		done <- fmt.Errorf("communication with %s timed out", addr)
	case err := <-commRv:
		done <- err
	}
}

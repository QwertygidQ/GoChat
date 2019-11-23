package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func send(conn *net.Conn, msg []byte) error {
	_, err := (*conn).Write(msg)
	if err != nil {
		fmt.Printf("Couldn't send the message \"%s\" -- %s", string(msg), err)
		return err
	}

	return nil
}

func read(conn *net.Conn) ([]byte, error) {
	buf := make([]byte, 1024)
	reqLen, err := (*conn).Read(buf)
	if err != nil {
		fmt.Printf("Couldn't read from the connection -- %s\n", err)
		return nil, err
	}

	return buf[:reqLen], nil
}

func readMessages(conn *net.Conn, deadChan chan<- bool) {
	for {
		buf, err := read(conn)
		if err != nil {
			(*conn).Close()
			deadChan <- true
			return
		}

		fmt.Println(string(buf))
	}
}

func sendMessages(conn *net.Conn, deadChan chan<- bool) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		if scanner.Scan() {
			msg := scanner.Text()

			err := send(conn, []byte(msg))
			if err != nil {
				(*conn).Close()
				deadChan <- true
				return
			}
		}
	}
}

func main() {
	args := os.Args[1:]
	if len(args) != 2 {
		fmt.Println("Usage: ./client ADDRESS USERNAME")
		return
	}

	conn, err := net.Dial("tcp", args[0])
	if err != nil {
		fmt.Printf("Failed to create the connection -- %s\n", err)
		return
	}

	err = send(&conn, []byte(args[1]))
	if err != nil {
		conn.Close()
		return
	}

	buf, err := read(&conn)
	if err != nil {
		conn.Close()
		return
	}

	fmt.Printf("Server message: %s\n", string(buf))

	deadChan := make(chan bool)
	go readMessages(&conn, deadChan)
	go sendMessages(&conn, deadChan)

	<-deadChan
	fmt.Println("Shutting down")
}

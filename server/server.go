package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
)

var users map[string]*net.Conn

func newUser(name string, conn *net.Conn) error {
	if _, ok := users[name]; ok {
		log.Printf("User %s already exists\n", name)
		return errors.New("User already exists")
	}

	users[name] = conn
	return nil
}

func removeUser(name string) error {
	if _, ok := users[name]; !ok {
		log.Printf("User %s is not present in the map\n", name)
		return errors.New("User does not exist in the map")
	}

	(*users[name]).Close()
	delete(users, name)

	broadcast("SERVER", []byte("User "+name+" has left the server"))
	return nil
}

func send(name string, msg []byte) error {
	_, err := (*users[name]).Write(msg)
	if err != nil {
		log.Printf("Couldn't send \"%s\" to user %s", string(msg), name)
		log.Println(err)
		return err
	}

	return nil
}

func broadcast(from string, msg []byte) {
	prefix := fmt.Sprintf("[%s] ", from)
	fullMsg := []byte(prefix)
	fullMsg = append(fullMsg, msg...)
	log.Printf("Broadcasting message from %s: \"%s\"", from, string(fullMsg))

	for name := range users {
		err := send(name, fullMsg)
		if err != nil {
			err := removeUser(name)
			if err != nil {
				log.Fatalf("Failed to remove user %s", name)
			}
		}
	}
}

func listenUser(name string) {
	conn := users[name]

	for {
		buf := make([]byte, 1024)
		reqLen, err := (*conn).Read(buf)
		if err != nil {
			log.Printf("Failed to read from an established connection: %s -- removing user %s\n", err, name)
			err := removeUser(name)
			if err != nil {
				log.Fatalf("Failed to remove user %s", name)
			}
			return
		}

		broadcast(name, buf[:reqLen])
	}
}

func createConnection(conn net.Conn) {
	buf := make([]byte, 1024)
	reqLen, err := conn.Read(buf)
	if err != nil {
		log.Printf("Failed to read from a new connection: %s\n", err)
		conn.Close()
		return
	}

	name := string(buf[:reqLen])
	err = newUser(name, &conn)
	if err != nil {
		log.Println(err)
		conn.Write([]byte("A user with this username already exists -- please, choose a different one"))
		conn.Close()
		return
	}

	_, err = conn.Write([]byte("Connected!"))
	if err != nil {
		conn.Close()
		return
	}

	broadcast("SERVER", []byte("User "+name+" has joined the server"))

	go listenUser(name)
	log.Printf("User %s created", name)
}

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Println("Usage: ./server ADDRESS")
		return
	}

	users = make(map[string]*net.Conn)

	ln, err := net.Listen("tcp", args[0])
	if err != nil {
		log.Fatal(err)
	}

	defer ln.Close()

	log.Println("Ready!")
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go createConnection(conn)
	}
}

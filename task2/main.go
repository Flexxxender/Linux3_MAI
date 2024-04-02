package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
)

const queueKey = "/tmp/queue"

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Not enough arguments")
		os.Exit(1)
	}

	file1, file2 := os.Args[1], os.Args[2]
	if file1 == "" || file2 == "" {
		fmt.Println("One of filenames is empty")
		os.Exit(1)
	}

	consumer, err := CreateConnection(queueKey)
	if err != nil {
		fmt.Println("Error to connect: ", err)
		os.Exit(1)
	}

	firstCmd := exec.Command("./bin/t1", file1)
	secondCmd := exec.Command("./bin/t1", file2)
	if err := firstCmd.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := secondCmd.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	firstMessage, err := consumer.ReadMessage()
	if err != nil {
		fmt.Println("Error reading message: ", err)
		os.Exit(1)
	}

	secondMessage, err := consumer.ReadMessage()
	if err != nil {
		fmt.Println("Error reading message: ", err)
		os.Exit(1)
	}

	ioutil.WriteFile("res.txt", xorBytes(firstMessage, secondMessage), fs.FileMode(0777))
}

func xorBytes(text, key []byte) []byte {
	textLen := len(text)
	dif := textLen - len(key)
	if dif > 0 {
		for i := 0; i < dif; i++ {
			key = append(key, 0)
		}
	}

	result := make([]byte, textLen)
	for i := 0; i < textLen; i++ {
		result[i] = key[i] ^ text[i]
	}

	return result
}

type MessageConsumer struct {
	connection *net.UnixConn
	address    string
}

type ChainMessage struct {
	LastInChain bool
	Content     []byte
}

func CreateConnection(address string) (*MessageConsumer, error) {
	os.Remove(address)
	addr, err := net.ResolveUnixAddr("unixgram", address)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUnixgram("unixgram", addr)
	if err != nil {
		return nil, err
	}

	return &MessageConsumer{connection: conn, address: address}, nil
}

func (c *MessageConsumer) ReadMessage() ([]byte, error) {
	buf := make([]byte, 1500)
	result := make([]byte, 0, 1024)
	isLastMessage := false
	for !isLastMessage {
		n, _, err := c.connection.ReadFromUnix(buf)
		if err != nil {
			return nil, err
		}

		bytes := buf[:n]
		var message ChainMessage
		err = json.Unmarshal(bytes, &message)
		if err != nil {
			return nil, err
		}
		result = append(result, message.Content...)
		isLastMessage = message.LastInChain
	}
	fmt.Printf("result: %v\n", len(result))
	return result, nil
}

func (c *MessageConsumer) CloseConnection() {
	os.Remove(c.address)
	c.connection.Close()
}

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
)

const (
	queueKey       = "/tmp/queue"
	bytesInMessage = 1024
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <filename>")
		os.Exit(1)
	}

	filePath := os.Args[1]

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file: ", err)
		os.Exit(1)
	}
	defer file.Close()

	body, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file: ", err)
		os.Exit(1)
	}
	queue, err := CreateConnection(queueKey)
	if err != nil {
		fmt.Println("Error while open connection: ", err)
		os.Exit(1)
	}
	defer queue.CloseConnection()

	for len(body) > bytesInMessage {
		buf := body[:bytesInMessage]
		message := ChainMessage{LastInChain: false, Content: buf}
		queue.SendMessage(message)
		body = body[bytesInMessage:]
	}

	message := ChainMessage{LastInChain: true, Content: body}
	queue.SendMessage(message)
}

type MessageQueue struct {
	connection *net.UnixConn
}

type ChainMessage struct {
	LastInChain bool
	Content     []byte
}

func CreateConnection(address string) (*MessageQueue, error) {
	addr, err := net.ResolveUnixAddr("unixgram", address)
	if err != nil {
		return nil, err
	}

	// Устанавливаем соединение с Unix сокетом
	conn, err := net.DialUnix("unixgram", nil, addr)
	if err != nil {
		return nil, err
	}

	return &MessageQueue{connection: conn}, nil
}

func (q *MessageQueue) SendMessage(message ChainMessage) error {
	messageBody, err := json.Marshal(message)
	if err != nil {
		return err
	}
	_, err = q.connection.Write(messageBody)
	return err
}

func (q *MessageQueue) CloseConnection() {
	q.connection.Close()
}

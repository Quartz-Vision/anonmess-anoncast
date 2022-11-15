package server

import (
	"anoncast/lists/squeue"
	"anoncast/settings"
	"errors"
	"io"
	"net"
	"sync"
)

var ErrServerStart = errors.New("server failed to start")
var ErrConnectionFailed = errors.New("connection hasn't been accepted")

var connections = sync.Map{}

func sendMessage(message []byte, conn net.Conn, waitGroup *sync.WaitGroup) {
	waitGroup.Add(1)
	defer waitGroup.Done()

	conn.Write(message)
}

func castMessages(excludeKey net.Addr, queue *squeue.SQueue, waitGroup *sync.WaitGroup) {
	waitGroup.Add(1)
	defer waitGroup.Done()

	for {
		for queue.IsEmpty() {
		}
		if val, ok := queue.Pop(); ok {
			if val == nil {
				break
			}

			valueWaitGroup := sync.WaitGroup{}
			connections.Range(func(key any, c any) bool {
				if key != excludeKey {
					go sendMessage(val.([]byte), c.(net.Conn), &valueWaitGroup)
				}
				return true
			})
			valueWaitGroup.Wait()
		}
	}
}

func handleConnection(conn net.Conn, waitGroup *sync.WaitGroup) (err error) {
	waitGroup.Add(1)
	defer conn.Close()
	defer waitGroup.Done()

	connKey := conn.RemoteAddr()
	connections.Store(connKey, conn)
	defer connections.Delete(connKey)

	var packageSize int32 = 0
	int32Buf := make([]byte, 4)

	messagesQueue := squeue.New()
	go castMessages(connKey, messagesQueue, waitGroup)
	defer messagesQueue.Push(nil)

	for {
		if _, err := io.ReadFull(conn, int32Buf); err != nil {
			return err
		}
		packageSize = BytesToInt32(int32Buf)
		packageBuf := make([]byte, packageSize)

		if _, err := io.ReadFull(conn, packageBuf); err != nil {
			return err
		}

		messagesQueue.Push(packageBuf)
	}
}

func Init() error {
	//establish connection
	listener, err := net.Listen("tcp", settings.Config.ServerAddr)

	if err != nil {
		return ErrServerStart
	}

	defer listener.Close()

	waitGroup := sync.WaitGroup{}
	var lastErr error = nil

	for {
		conn, err := listener.Accept()
		if err != nil {
			lastErr = ErrConnectionFailed
			break
		}

		go handleConnection(conn, &waitGroup)
	}

	waitGroup.Wait()
	return lastErr
}

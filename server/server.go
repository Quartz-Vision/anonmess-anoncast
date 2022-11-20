package server

import (
	"anoncast/lists/squeue"
	"anoncast/settings"
	"errors"
	"io"
	"net"
	"sync"
	"time"
)

var ErrServerStart = errors.New("server failed to start")
var ErrConnectionFailed = errors.New("connection hasn't been accepted")

var connections = sync.Map{}

func sendMessage(message *[]byte, conn net.Conn, waitGroup *sync.WaitGroup) {
	conn.Write(*message)
	waitGroup.Done()
}

func castMessages(excludeKey net.Addr, queue *squeue.SQueue, waitGroup *sync.WaitGroup) {
root:
	for {
		for queue.IsEmpty() {
			time.Sleep(time.Millisecond)
		}

		for val, ok := queue.Pop(); ok; val, ok = queue.Pop() {
			if val == nil {
				break root
			}

			valueWaitGroup := sync.WaitGroup{}
			connections.Range(func(key any, c any) bool {
				if key != excludeKey {
					valueWaitGroup.Add(1)
					go sendMessage(val.(*[]byte), c.(net.Conn), &valueWaitGroup)
				}
				return true
			})
			valueWaitGroup.Wait()
		}
	}

	waitGroup.Done()
}

func handleConnection(conn net.Conn, waitGroup *sync.WaitGroup) (err error) {
	connKey := conn.RemoteAddr()
	connections.Store(connKey, conn)

	messagesQueue := squeue.New()
	castWaitGroup := sync.WaitGroup{}
	castWaitGroup.Add(1)
	go castMessages(connKey, messagesQueue, &castWaitGroup)

	var retErr error = nil

	for {
		sizeRawBuf := make([]byte, INT64_SIZE)

		if _, err := io.ReadFull(conn, sizeRawBuf); err != nil {
			retErr = err
			break
		}

		packageSize, _ := BytesToInt64(sizeRawBuf)
		sizeBufLen := int64(len(sizeRawBuf))
		packageBuf := make([]byte, packageSize+sizeBufLen)

		copy(packageBuf, sizeRawBuf)

		if _, err := io.ReadFull(conn, packageBuf[sizeBufLen:]); err != nil {
			retErr = err
			break
		}

		messagesQueue.Push(&packageBuf)
	}

	// Defers are too slow for this part
	messagesQueue.Push(nil)
	castWaitGroup.Wait()
	conn.Close()
	connections.Delete(connKey)
	waitGroup.Done()

	return retErr
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

		waitGroup.Add(1)
		go handleConnection(conn, &waitGroup)
	}

	waitGroup.Wait()
	return lastErr
}

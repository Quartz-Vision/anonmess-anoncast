package server

import (
	"anoncast/settings"
	"context"
	"errors"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"

	"github.com/Quartz-Vision/golog"
)

const (
	MAX_PACKAGE_SIZE_B = 1 << 20 // 1MB

	// The last byte in a package to check for validity.
	// Doesn't ensure 100% correctness, but eliminates some stupid cases.
	CHECK_SYMBOL_B byte = 0b10110010
)

var (
	ErrServerStart      = errors.New("server failed to start")
	ErrConnectionFailed = errors.New("connection hasn't been accepted")
	connections         = sync.Map{}
)

func sendMessage(message *[]byte, conn net.Conn, waitGroup *sync.WaitGroup) {
	if _, err := conn.Write(*message); err != nil && err != io.EOF {
		golog.Warning.Println("Package writing failed, skipping connection", err.Error())
	}
	waitGroup.Done()
}

func castMessages(excludeKey net.Addr, queue chan *[]byte, waitGroup *sync.WaitGroup) {
	for message := range queue {
		valueWaitGroup := sync.WaitGroup{}
		connections.Range(func(key any, c any) bool {
			if key != excludeKey {
				valueWaitGroup.Add(1)
				go sendMessage(message, c.(net.Conn), &valueWaitGroup)
			}
			return true
		})
		valueWaitGroup.Wait()
	}

	waitGroup.Done()
}

func handleConnection(conn net.Conn, waitGroup *sync.WaitGroup) error {
	connKey := conn.RemoteAddr()
	connections.Store(connKey, conn)

	messages := make(chan *[]byte, 8)
	castWaitGroup := sync.WaitGroup{}
	castWaitGroup.Add(1)
	go castMessages(connKey, messages, &castWaitGroup)

	var retErr error = nil

	for {
		sizeRawBuf := make([]byte, INT64_SIZE)

		if _, err := io.ReadFull(conn, sizeRawBuf); err != nil {
			if err != io.EOF {
				golog.Warning.Println("Package reading failed, dropping connection", err.Error())
				retErr = err
			}
			break
		}

		packageSize, sizeBufLen := BytesToInt64(sizeRawBuf)
		if packageSize <= 0 || packageSize >= MAX_PACKAGE_SIZE_B {
			golog.Warning.Printf("Received wrong package size (%v), dropping connection", packageSize)
			break
		}
		packageBuf := make([]byte, packageSize+int64(sizeBufLen))

		copy(packageBuf, sizeRawBuf)

		if _, err := io.ReadFull(conn, packageBuf[sizeBufLen:]); err != nil || packageBuf[packageSize-1] != CHECK_SYMBOL_B {
			if err == io.ErrUnexpectedEOF {
				golog.Warning.Println("Package reading failed, dropping connection", err.Error())
				retErr = err
			}
			break
		}

		messages <- &packageBuf
	}

	// Defers are too slow for this part
	close(messages)
	castWaitGroup.Wait()
	conn.Close()
	connections.Delete(connKey)
	waitGroup.Done()

	return retErr
}

func onInterruption(ctx context.Context, callback func(os.Signal)) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, os.Kill)

	select {
	case sig := <-sigChan:
		callback(sig)
	case <-ctx.Done():
		close(sigChan)
	}
}

func Start() error {
	waitGroup := sync.WaitGroup{}
	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	listener, err := net.Listen("tcp", settings.Config.ServerAddr)
	defer listener.Close()

	go onInterruption(ctx, func(s os.Signal) {
		golog.Info.Println("Process interrupted")
		stop()
		listener.Close()
	})

	if err != nil {
		golog.Error.Println(err.Error())
		return ErrServerStart
	}

serverLoop:
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				break serverLoop
			default:
				golog.Error.Println(err.Error())
				continue serverLoop
			}
		}
		waitGroup.Add(1)
		go handleConnection(conn, &waitGroup)
	}

	waitGroup.Wait()
	return nil
}

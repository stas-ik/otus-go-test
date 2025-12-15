package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	var timeout time.Duration
	flag.DurationVar(&timeout, "timeout", 10*time.Second, "connection timeout")
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: go-telnet [--timeout=<duration>] host port\n")
		os.Exit(1)
	}

	host := args[0]
	port := args[1]
	address := net.JoinHostPort(host, port)

	client := NewTelnetClient(address, timeout, os.Stdin, os.Stdout)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to %s: %v\n", address, err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Fprintf(os.Stderr, "...Connected to %s\n", address)

	var wg sync.WaitGroup
	wg.Add(2)

	// горутина для отправки данных из STDIN в сокет
	go func() {
		defer wg.Done()
		defer cancel() // сигнализируем о завершении

		for {
			select {
			case <-ctx.Done():
				return
			default:
				err := client.Send()
				if err == io.EOF {
					fmt.Fprintf(os.Stderr, "...EOF\n")
					return
				}
				if err != nil {
					// проверяем, не закрыто ли соединение
					if isConnectionClosed(err) {
						fmt.Fprintf(os.Stderr, "...Connection was closed by peer\n")
						return
					}
					fmt.Fprintf(os.Stderr, "Send error: %v\n", err)
					return
				}
			}
		}
	}()

	//горутина для получения данных из сокета в STDOUT
	go func() {
		defer wg.Done()
		defer cancel() // сигнализируем о завершении

		for {
			select {
			case <-ctx.Done():
				return
			default:
				err := client.Receive()
				if err == io.EOF {
					fmt.Fprintf(os.Stderr, "...Connection was closed by peer\n")
					return
				}
				if err != nil {
					if isConnectionClosed(err) {
						fmt.Fprintf(os.Stderr, "...Connection was closed by peer\n")
						return
					}
					fmt.Fprintf(os.Stderr, "Receive error: %v\n", err)
					return
				}
			}
		}
	}()

	<-ctx.Done()
	wg.Wait()
}

func isConnectionClosed(err error) bool {
	if err == nil {
		return false
	}

	switch err {
	case io.EOF, io.ErrClosedPipe:
		return true
	}

	//проверяем сетевые ошибки
	if netErr, ok := err.(*net.OpError); ok {
		if netErr.Err.Error() == "use of closed network connection" {
			return true
		}
	}

	return false
}

package main

import (
	"context"
	"errors"
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
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var timeout time.Duration
	flag.DurationVar(&timeout, "timeout", 10*time.Second, "connection timeout")
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		return fmt.Errorf("usage: %s [--timeout=duration] host port", os.Args[0])
	}

	host := args[0]
	port := args[1]
	address := net.JoinHostPort(host, port)

	// создаем контекст с обработкой сигналов
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer cancel()

	client := NewTelnetClient(address, timeout, os.Stdin, os.Stdout)

	if err := client.Connect(); err != nil {
		return fmt.Errorf("connection failed to %s: %w", address, err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Close error: %v\n", closeErr)
		}
	}()

	fmt.Fprintf(os.Stderr, "...Connected to %s\n", address)

	runConcurrentIO(ctx, client, cancel)
	return nil
}

func runConcurrentIO(ctx context.Context, client TelnetClient, cancel context.CancelFunc) {
	var wg sync.WaitGroup
	wg.Add(2)

	// горутина для отправки данных из стдин в сокет
	go func() {
		defer wg.Done()
		defer cancel()

		if err := client.Send(); err != nil {
			handleSendError(err)
		}
	}()

	// горутина для получения данных из сокета в стдаут
	go func() {
		defer wg.Done()
		defer cancel()

		if err := client.Receive(); err != nil {
			handleReceiveError(err)
		}
	}()

	// ждем сигнал или завершение горутин
	<-ctx.Done()

	// закрываем соединение, чтобы завершить горутины
	if closeErr := client.Close(); closeErr != nil {
		fmt.Fprintf(os.Stderr, "Close error during shutdown: %v\n", closeErr)
	}

	wg.Wait()
}

func handleSendError(err error) {
	switch {
	case errors.Is(err, io.EOF):
		fmt.Fprintf(os.Stderr, "...EOF\n")
	case isConnectionClosed(err):
		fmt.Fprintf(os.Stderr, "...Connection was closed by peer\n")
	default:
		fmt.Fprintf(os.Stderr, "Send error: %v\n", err)
	}
}

func handleReceiveError(err error) {
	switch {
	case errors.Is(err, io.EOF):
		fmt.Fprintf(os.Stderr, "...Connection was closed by peer\n")
	case isConnectionClosed(err):
		fmt.Fprintf(os.Stderr, "...Connection was closed by peer\n")
	default:
		fmt.Fprintf(os.Stderr, "Receive error: %v\n", err)
	}
}

func isConnectionClosed(err error) bool {
	if err == nil {
		return false
	}

	// проверяем различные типы ошибок, которые могут означать закрытое соединение
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) {
		return true
	}

	// проверяем сетевые ошибки
	var netErr *net.OpError
	if errors.As(err, &netErr) && netErr.Err.Error() == "use of closed network connection" {
		return true
	}

	return false
}

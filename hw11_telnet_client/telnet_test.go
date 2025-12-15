package main

import (
	"bytes"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTelnetClient(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()

			in := &bytes.Buffer{}
			out := &bytes.Buffer{}

			timeout, err := time.ParseDuration("10s")
			require.NoError(t, err)

			client := NewTelnetClient(l.Addr().String(), timeout, io.NopCloser(in), out)
			require.NoError(t, client.Connect())
			defer func() { require.NoError(t, client.Close()) }()

			in.WriteString("hello\n")
			err = client.Send()
			require.NoError(t, err)

			err = client.Receive()
			require.NoError(t, err)
			require.Equal(t, "world\n", out.String())
		}()

		go func() {
			defer wg.Done()

			conn, err := l.Accept()
			require.NoError(t, err)
			require.NotNil(t, conn)
			defer func() { require.NoError(t, conn.Close()) }()

			request := make([]byte, 1024)
			n, err := conn.Read(request)
			require.NoError(t, err)
			require.Equal(t, "hello\n", string(request)[:n])

			n, err = conn.Write([]byte("world\n"))
			require.NoError(t, err)
			require.NotEqual(t, 0, n)
		}()

		wg.Wait()
	})

	t.Run("connection timeout", func(t *testing.T) {
		timeout := 1 * time.Second
		client := NewTelnetClient("192.0.2.1:12345", timeout, io.NopCloser(&bytes.Buffer{}), &bytes.Buffer{})

		start := time.Now()
		err := client.Connect()
		elapsed := time.Since(start)

		require.Error(t, err)
		require.Less(t, elapsed, timeout+500*time.Millisecond)
	})

	t.Run("send and receive", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		// Сервер
		go func() {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			defer conn.Close()

			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil {
				return
			}

			response := "echo: " + string(buf[:n])
			conn.Write([]byte(response))
		}()

		in := bytes.NewBufferString("test message\n")
		out := &bytes.Buffer{}

		client := NewTelnetClient(l.Addr().String(), 5*time.Second, io.NopCloser(in), out)
		require.NoError(t, client.Connect())
		defer client.Close()

		require.NoError(t, client.Send())

		time.Sleep(100 * time.Millisecond)
		require.NoError(t, client.Receive())

		require.Contains(t, out.String(), "echo: test message")
	})
}

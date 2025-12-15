package main

import (
	"bufio"
	"io"
	"net"
	"time"
)

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

type telnetClient struct {
	address string
	timeout time.Duration
	in      io.ReadCloser
	out     io.Writer
	conn    net.Conn
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &telnetClient{
		address: address,
		timeout: timeout,
		in:      in,
		out:     out,
	}
}

func (tc *telnetClient) Connect() error {
	conn, err := net.DialTimeout("tcp", tc.address, tc.timeout)
	if err != nil {
		return err
	}
	tc.conn = conn
	return nil
}

func (tc *telnetClient) Close() error {
	if tc.conn != nil {
		return tc.conn.Close()
	}
	return nil
}

func (tc *telnetClient) Send() error {
	if tc.conn == nil {
		return io.ErrClosedPipe
	}

	scanner := bufio.NewScanner(tc.in)
	if scanner.Scan() {
		line := scanner.Text() + "\n"
		_, err := tc.conn.Write([]byte(line))
		return err
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return io.EOF
}

func (tc *telnetClient) Receive() error {
	if tc.conn == nil {
		return io.ErrClosedPipe
	}

	reader := bufio.NewReader(tc.conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	_, err = tc.out.Write([]byte(line))
	return err
}

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
	return &telnetClient{address: address, timeout: timeout, in: in, out: out}
}

func (c *telnetClient) Connect() error {
	d := net.Dialer{Timeout: c.timeout}
	conn, err := d.Dial("tcp", c.address)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *telnetClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *telnetClient) Send() error {
	if c.conn == nil {
		return io.ErrClosedPipe
	}
	r := bufio.NewReader(c.in)
	b, err := r.ReadBytes('\n')
	if len(b) > 0 {
		if _, werr := c.conn.Write(b); werr != nil {
			return werr
		}
	}
	if err != nil && err != io.EOF {
		return err
	}
	if err == io.EOF && len(b) == 0 {
		return io.EOF
	}
	return nil
}

func (c *telnetClient) Receive() error {
	if c.conn == nil {
		return io.ErrClosedPipe
	}
	r := bufio.NewReader(c.conn)
	b, err := r.ReadBytes('\n')
	if len(b) > 0 {
		if _, werr := c.out.Write(b); werr != nil {
			return werr
		}
	}
	if err != nil && err != io.EOF {
		return err
	}
	if err == io.EOF && len(b) == 0 {
		return io.EOF
	}
	return nil
}

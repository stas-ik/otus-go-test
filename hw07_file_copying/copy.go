package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

type progressWriter struct {
	dst     io.Writer
	total   int64
	written int64
	lastPct int
}

func newProgressWriter(dst io.Writer, total int64) *progressWriter {
	return &progressWriter{
		dst:     dst,
		total:   total,
		lastPct: -1,
	}
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.dst.Write(p)
	pw.written += int64(n)
	pw.printProgress()
	return n, err
}

func (pw *progressWriter) printProgress() {
	if pw.total <= 0 {
		return
	}
	pct := int(float64(pw.written) / float64(pw.total) * 100)
	if pct > 100 {
		pct = 100
	}
	if pct > pw.lastPct {
		fmt.Printf("%d%%\n", pct)
		pw.lastPct = pct
	}
}

func Copy(fromPath, toPath string, offset, limit int64) error {
	src, err := os.Open(fromPath)
	if err != nil {
		return err
	}
	defer src.Close()

	fi, err := src.Stat()
	if err != nil {
		return err
	}
	if !fi.Mode().IsRegular() {
		return ErrUnsupportedFile
	}
	fileSize := fi.Size()

	if offset > fileSize {
		return ErrOffsetExceedsFileSize
	}

	dst, err := os.Create(toPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = dst.Close()
	}()

	remaining := fileSize - offset
	toCopy := limit
	if limit == 0 || limit > remaining {
		toCopy = remaining
	}

	if toCopy == 0 {
		fmt.Println("100%")
		return nil
	}

	section := io.NewSectionReader(src, offset, toCopy)

	pw := newProgressWriter(dst, toCopy)

	buf := make([]byte, 32*1024)
	if _, err := io.CopyBuffer(pw, section, buf); err != nil {
		return err
	}

	if pw.lastPct < 100 {
		fmt.Println("100%")
	}

	return nil
}

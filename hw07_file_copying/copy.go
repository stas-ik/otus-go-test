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

	if _, err := src.Seek(offset, io.SeekStart); err != nil {
		return err
	}

	remaining := fileSize - offset
	toCopy := limit
	if limit == 0 || limit > remaining {
		toCopy = remaining
	}

	if toCopy == 0 {
		fmt.Println("100%")
		return nil
	}

	buf := make([]byte, 32*1024)
	var copied int64
	for copied < toCopy {
		chunk := toCopy - copied
		if int64(len(buf)) < chunk {
			chunk = int64(len(buf))
		}
		n, rerr := io.ReadFull(src, buf[:chunk])
		if rerr != nil {
			if rerr == io.EOF || errors.Is(rerr, io.ErrUnexpectedEOF) {
				if n > 0 {
					wn, werr := dst.Write(buf[:n])
					copied += int64(wn)
					if werr != nil {
						return werr
					}
				}
				break
			}
			return rerr
		}
		wn, werr := dst.Write(buf[:n])
		copied += int64(wn)
		if werr != nil {
			return werr
		}
		pct := int(float64(copied) / float64(toCopy) * 100)
		fmt.Printf("%d%%\n", pct)
	}

	if copied >= toCopy {
		fmt.Println("100%")
	}
	return nil
}

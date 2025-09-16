package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCopy(t *testing.T) {
	td := "testdata"
	in := filepath.Join(td, "input.txt")

	t.Run("offset0_limit0", func(t *testing.T) {
		wantPath := filepath.Join(td, "out_offset0_limit0.txt")
		out := tempFile(t)
		defer os.Remove(out)
		if err := Copy(in, out, 0, 0); err != nil {
			t.Fatalf("Copy error: %v", err)
		}
		compareFiles(t, out, wantPath)
	})

	t.Run("offset0_limit10", func(t *testing.T) {
		wantPath := filepath.Join(td, "out_offset0_limit10.txt")
		out := tempFile(t)
		defer os.Remove(out)
		if err := Copy(in, out, 0, 10); err != nil {
			t.Fatalf("Copy error: %v", err)
		}
		compareFiles(t, out, wantPath)
	})

	t.Run("offset0_limit1000", func(t *testing.T) {
		wantPath := filepath.Join(td, "out_offset0_limit1000.txt")
		out := tempFile(t)
		defer os.Remove(out)
		if err := Copy(in, out, 0, 1000); err != nil {
			t.Fatalf("Copy error: %v", err)
		}
		compareFiles(t, out, wantPath)
	})

	t.Run("offset0_limit10000", func(t *testing.T) {
		wantPath := filepath.Join(td, "out_offset0_limit10000.txt")
		out := tempFile(t)
		defer os.Remove(out)
		if err := Copy(in, out, 0, 10000); err != nil {
			t.Fatalf("Copy error: %v", err)
		}
		compareFiles(t, out, wantPath)
	})

	t.Run("offset100_limit1000", func(t *testing.T) {
		wantPath := filepath.Join(td, "out_offset100_limit1000.txt")
		out := tempFile(t)
		defer os.Remove(out)
		if err := Copy(in, out, 100, 1000); err != nil {
			t.Fatalf("Copy error: %v", err)
		}
		compareFiles(t, out, wantPath)
	})

	t.Run("offset6000_limit1000", func(t *testing.T) {
		wantPath := filepath.Join(td, "out_offset6000_limit1000.txt")
		out := tempFile(t)
		defer os.Remove(out)
		if err := Copy(in, out, 6000, 1000); err != nil {
			t.Fatalf("Copy error: %v", err)
		}
		compareFiles(t, out, wantPath)
	})

	t.Run("offset_exceeds_size", func(t *testing.T) {
		out := tempFile(t)
		defer os.Remove(out)
		err := Copy(in, out, 1<<60, 10)
		if !errors.Is(err, ErrOffsetExceedsFileSize) {
			t.Fatalf("expected ErrOffsetExceedsFileSize, got %v", err)
		}
	})

	t.Run("unsupported_file", func(t *testing.T) {
		dir := td
		out := tempFile(t)
		defer os.Remove(out)
		err := Copy(dir, out, 0, 10)
		if !errors.Is(err, ErrUnsupportedFile) {
			t.Fatalf("expected ErrUnsupportedFile, got %v", err)
		}
	})
}

func tempFile(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp("", "copytest-*.out")
	if err != nil {
		t.Fatalf("TempFile: %v", err)
	}
	name := f.Name()
	_ = f.Close()
	return name
}

func compareFiles(t *testing.T, gotPath, wantPath string) {
	t.Helper()
	got, err := os.ReadFile(gotPath)
	if err != nil {
		t.Fatalf("read got: %v", err)
	}
	want, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("read want: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("files are different: got %s want %s", gotPath, wantPath)
	}
}

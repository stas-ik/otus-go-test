//go:build bench
// +build bench

package hw10programoptimization

import (
	"archive/zip"
	"testing"
)

// BenchmarkGetDomainStat runs GetDomainStat on the provided dataset and reports time and allocations.
// Usage:
//
//	go test -bench=BenchmarkGetDomainStat -benchmem -tags bench ./hw10_program_optimization
//
// For multiple runs suitable for benchstat:
//
//	go test -bench=BenchmarkGetDomainStat -benchmem -count=10 -tags bench ./hw10_program_optimization > before.txt
//	# make your changes
//	go test -bench=BenchmarkGetDomainStat -benchmem -count=10 -tags bench ./hw10_program_optimization > after.txt
//	benchstat before.txt after.txt
func BenchmarkGetDomainStat(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r, err := zip.OpenReader("testdata/users.dat.zip")
		if err != nil {
			b.Fatalf("open zip: %v", err)
		}
		if len(r.File) != 1 {
			r.Close()
			b.Fatalf("unexpected files in archive: %d", len(r.File))
		}
		data, err := r.File[0].Open()
		if err != nil {
			r.Close()
			b.Fatalf("open file in zip: %v", err)
		}

		stat, err := GetDomainStat(data, "biz")
		if err != nil {
			data.Close()
			r.Close()
			b.Fatalf("GetDomainStat error: %v", err)
		}
		// validate to ensure we're benchmarking the correct work
		if len(stat) != len(expectedBizStat) {
			data.Close()
			r.Close()
			b.Fatalf("unexpected stat size: got %d, want %d", len(stat), len(expectedBizStat))
		}
		for k, v := range expectedBizStat {
			if stat[k] != v {
				data.Close()
				r.Close()
				b.Fatalf("unexpected value for %s: got %d, want %d", k, stat[k], v)
			}
		}

		data.Close()
		r.Close()
	}
}

package hw10programoptimization

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type User struct {
	ID       int
	Name     string
	Username string
	Email    string
	Phone    string
	Password string
	Address  string
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	if r == nil {
		return nil, fmt.Errorf("nil reader")
	}
	ld := strings.ToLower(domain)
	if ld == "" {
		return DomainStat{}, nil
	}

	scanner := bufio.NewScanner(r)
	buf := make([]byte, 0, 1024*64)
	scanner.Buffer(buf, 1024*1024)

	result := make(DomainStat)
	var rec struct {
		Email string `json:"email"`
	}

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		rec.Email = ""
		if err := json.Unmarshal(line, &rec); err != nil {
			return nil, fmt.Errorf("unmarshal error: %w", err)
		}
		if rec.Email == "" {
			continue
		}
		email := strings.ToLower(rec.Email)
		at := strings.LastIndexByte(email, '@')
		if at == -1 || at+1 >= len(email) {
			continue
		}
		dom := email[at+1:]
		if strings.HasSuffix(dom, "."+ld) {
			result[dom]++
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan error: %w", err)
	}
	return result, nil
}

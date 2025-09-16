package hw09structvalidator

import (
	"errors"
	"testing"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int      `validate:"min:18|max:50"`
		Email  string   `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole `validate:"in:admin,stuff"`
		Phones []string `validate:"len:11"`
		meta   []byte   //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}
)

func TestValidate(t *testing.T) {
	t.Run("valid user passes", func(t *testing.T) {
		u := User{
			ID:     "12345678-1234-1234-1234-123456789012", // 36 //nolint:lll // test data UUID is long
			Name:   "John",
			Age:    30,
			Email:  "john@doe.com",
			Role:   "admin",
			Phones: []string{"12345678901", "10987654321"},
		}
		if err := Validate(u); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("invalid user accumulates errors", func(t *testing.T) {
		u := User{
			ID:     "short",                        // len fail
			Age:    10,                             // min fail
			Email:  "not-email",                    // regex fail
			Role:   "guest",                        // in fail
			Phones: []string{"123", "45678901234"}, // first fails len
		}
		err := Validate(u)
		var verrs ValidationErrors
		if !errors.As(err, &verrs) {
			t.Fatalf("expected ValidationErrors, got %v", err)
		}
		if len(verrs) < 5 {
			t.Fatalf("expected >=5 validation errors, got %d (%v)", len(verrs), err)
		}
		foundLen := false
		foundMin := false
		foundRegexp := false
		foundIn := false
		for _, ve := range verrs {
			if errors.Is(ve.Err, ErrRuleLen) {
				foundLen = true
			}
			if errors.Is(ve.Err, ErrRuleMin) {
				foundMin = true
			}
			if errors.Is(ve.Err, ErrRuleRegexp) {
				foundRegexp = true
			}
			if errors.Is(ve.Err, ErrRuleIn) {
				foundIn = true
			}
		}
		if !(foundLen && foundMin && foundRegexp && foundIn) {
			t.Fatalf("missing expected rule errors: len=%v min=%v regexp=%v in=%v", foundLen, foundMin, foundRegexp, foundIn)
		}
	})

	t.Run("non-struct gives programming error", func(t *testing.T) {
		if err := Validate(123); !errors.Is(err, ErrNotStruct) {
			t.Fatalf("expected ErrNotStruct, got %v", err)
		}
	})

	t.Run("unsupported field type with validate returns programming error", func(t *testing.T) {
		type Bad struct {
			B []byte `validate:"len:3"`
		}
		err := Validate(Bad{B: []byte{1, 2, 3}})
		if !errors.Is(err, ErrUnsupported) {
			t.Fatalf("expected ErrUnsupported, got %v", err)
		}
	})

	t.Run("bad tag name returns programming error", func(t *testing.T) {
		type S struct {
			A string `validate:"unknown:1"`
		}
		err := Validate(S{A: "x"})
		if !errors.Is(err, ErrInvalidTag) {
			t.Fatalf("expected ErrInvalidTag, got %v", err)
		}
	})

	t.Run("bad regexp returns compile error", func(t *testing.T) {
		type S struct {
			A string `validate:"regexp:("`
		}
		err := Validate(S{A: "any"})
		if err == nil {
			t.Fatalf("expected error")
		}
		// not a ValidationErrors
		var verrs ValidationErrors
		if errors.As(err, &verrs) {
			t.Fatalf("expected programming error, got validation errors")
		}
	})

	t.Run("App version len", func(t *testing.T) {
		if err := Validate(App{Version: "1.2.3"}); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if err := Validate(App{Version: "1.2"}); err == nil {
			t.Fatalf("expected validation error")
		}
	})

	t.Run("Response code in", func(t *testing.T) {
		if err := Validate(Response{Code: 200}); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if err := Validate(Response{Code: 201}); err == nil {
			t.Fatalf("expected validation error")
		}
	})
}

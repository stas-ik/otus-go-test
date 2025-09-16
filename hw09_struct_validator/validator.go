// Package hw09structvalidator provides a simple struct validation utility
// that validates exported struct fields according to "validate" tags.
package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// ValidationError represents a validation failure for a specific field.
// Field contains the field (optionally with index for slices) and Err holds the rule error.
type ValidationError struct {
	Field string
	Err   error
}

// ValidationErrors is a collection of ValidationError that implements the error interface.
type ValidationErrors []ValidationError

// Error implements the error interface for ValidationErrors.
func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return "no validation errors"
	}
	var b strings.Builder
	for i, ve := range v {
		if i > 0 {
			b.WriteString("; ")
		}
		b.WriteString(fmt.Sprintf("%s: %v", ve.Field, ve.Err))
	}
	return b.String()
}

// Exposed error variables describing validation and programming errors.
// They are used both for signaling programming/config errors and for matching rule failures.
var (
	// ErrNotStruct is returned when the input to Validate is not a struct (or a non-nil pointer to struct).
	ErrNotStruct = errors.New("input is not a struct")
	// ErrUnsupported indicates a field type is not supported by the validator for the provided rules.
	ErrUnsupported = errors.New("unsupported type for validation")
	// ErrInvalidTag indicates an invalid validate tag or parameter.
	ErrInvalidTag = errors.New("invalid validate tag")
	// ErrRuleLen indicates the string length rule has failed.
	ErrRuleLen = errors.New("len rule failed")
	// ErrRuleRegexp indicates the regexp rule has failed.
	ErrRuleRegexp = errors.New("regexp rule failed")
	// ErrRuleIn indicates the inclusion (in) rule has failed.
	ErrRuleIn = errors.New("in rule failed")
	// ErrRuleMin indicates the minimum numeric rule has failed.
	ErrRuleMin = errors.New("min rule failed")
	// ErrRuleMax indicates the maximum numeric rule has failed.
	ErrRuleMax = errors.New("max rule failed")
)

type rule struct {
	name  string
	param string
}

func parseTag(tag string) ([]rule, error) {
	if tag == "" {
		return nil, nil
	}
	raw := strings.Split(tag, "|")
	rules := make([]rule, 0, len(raw))
	for _, part := range raw {
		kv := strings.SplitN(strings.TrimSpace(part), ":", 2)
		if len(kv) != 2 {
			return nil, ErrInvalidTag
		}
		rules = append(rules, rule{name: kv[0], param: kv[1]})
	}
	return rules, nil
}

// Validate validates exported fields of a struct according to the `validate` struct tag.
// It returns ValidationErrors for user data errors, or a direct error for programming/tag errors.
func Validate(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return ErrNotStruct
		}
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return ErrNotStruct
	}
	typ := val.Type()
	var verrs ValidationErrors
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}
		tag := field.Tag.Get("validate")
		if tag == "" {
			continue
		}
		rules, err := parseTag(tag)
		if err != nil {
			return err
		}
		fv := val.Field(i)
		if fv.Kind() == reflect.Slice {
			for j := 0; j < fv.Len(); j++ {
				if err := validateValue(field.Name, fv.Index(j), rules, j, &verrs); err != nil {
					return err
				}
			}
			continue
		}
		if err := validateValue(field.Name, fv, rules, -1, &verrs); err != nil {
			return err
		}
	}
	if len(verrs) > 0 {
		return verrs
	}
	return nil
}

func validateValue(fieldName string, v reflect.Value, rules []rule, idx int, verrs *ValidationErrors) error {
	kind := v.Kind()
	for _, r := range rules {
		switch kind {
		case reflect.String:
			if err := applyStringRule(v.String(), r); err != nil {
				if isProgrammingError(err) {
					return err
				}
				addVE(verrs, fieldName, idx, err)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if err := applyIntRule(v.Int(), r); err != nil {
				if isProgrammingError(err) {
					return err
				}
				addVE(verrs, fieldName, idx, err)
			}
		default:
			return ErrUnsupported
		}
	}
	return nil
}

func addVE(verrs *ValidationErrors, field string, idx int, err error) {
	name := field
	if idx >= 0 {
		name = fmt.Sprintf("%s[%d]", field, idx)
	}
	*verrs = append(*verrs, ValidationError{Field: name, Err: err})
}

func applyStringRule(s string, r rule) error {
	switch r.name {
	case "len":
		want, err := strconv.Atoi(r.param)
		if err != nil {
			return ErrInvalidTag
		}
		if len(s) != want {
			return fmt.Errorf("%w: want %d got %d", ErrRuleLen, want, len(s))
		}
	case "regexp":
		re, err := regexp.Compile(r.param)
		if err != nil {
			return err
		}
		if !re.MatchString(s) {
			return fmt.Errorf("%w: %s", ErrRuleRegexp, r.param)
		}
	case "in":
		options := strings.Split(r.param, ",")
		for i := range options {
			options[i] = strings.TrimSpace(options[i])
		}
		for _, opt := range options {
			if s == opt {
				return nil
			}
		}
		return fmt.Errorf("%w: %s", ErrRuleIn, r.param)
	default:
		return ErrInvalidTag
	}
	return nil
}

func isProgrammingError(err error) bool {
	if errors.Is(err, ErrInvalidTag) {
		return true
	}
	ruleErrors := []error{
		ErrRuleLen,
		ErrRuleRegexp,
		ErrRuleIn,
		ErrRuleMin,
		ErrRuleMax,
	}
	for _, re := range ruleErrors {
		if errors.Is(err, re) {
			return false
		}
	}
	return true
}

func applyIntRule(n int64, r rule) error {
	switch r.name {
	case "min":
		v, err := strconv.ParseInt(r.param, 10, 64)
		if err != nil {
			return ErrInvalidTag
		}
		if n < v {
			return fmt.Errorf("%w: %d", ErrRuleMin, v)
		}
	case "max":
		v, err := strconv.ParseInt(r.param, 10, 64)
		if err != nil {
			return ErrInvalidTag
		}
		if n > v {
			return fmt.Errorf("%w: %d", ErrRuleMax, v)
		}
	case "in":
		parts := strings.Split(r.param, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			v, err := strconv.ParseInt(p, 10, 64)
			if err != nil {
				return ErrInvalidTag
			}
			if n == v {
				return nil
			}
		}
		return fmt.Errorf("%w: %s", ErrRuleIn, r.param)
	default:
		return ErrInvalidTag
	}
	return nil
}

//go:build !bench
// +build !bench

package hw10programoptimization

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDomainStat_EmptyInput(t *testing.T) {
	res, err := GetDomainStat(bytes.NewBuffer(nil), "com")
	require.NoError(t, err)
	require.Empty(t, res)
}

func TestGetDomainStat_InvalidJSONLine(t *testing.T) {
	data := `{"Id":1,"Email":"user@Example.COM"}
{invalid json line}`
	_, err := GetDomainStat(bytes.NewBufferString(data), "com")
	require.Error(t, err)
}

func TestGetDomainStat_MissingEmailField(t *testing.T) {
	data := `{"Id":1,"Name":"No Email"}
{"Id":2,"Email":"user@Example.COM"}`
	res, err := GetDomainStat(bytes.NewBufferString(data), "com")
	require.NoError(t, err)
	require.Equal(t, DomainStat{"example.com": 1}, res)
}

func TestGetDomainStat_CaseInsensitiveTLDMatch(t *testing.T) {
	data := `{"Email":"a@Foo.GOV"}
{"Email":"b@Bar.gov"}
{"Email":"c@baz.Gov"}`
	res, err := GetDomainStat(bytes.NewBufferString(data), "gov")
	require.NoError(t, err)
	require.Equal(t, DomainStat{"foo.gov": 1, "bar.gov": 1, "baz.gov": 1}, res)
}

func TestGetDomainStat_DomainFilterNotSubstringOfName(t *testing.T) {
	data := `{"Email":"co.m-user@site.xyz"}
{"Email":"user@com.net"}`
	res, err := GetDomainStat(bytes.NewBufferString(data), "com")
	require.NoError(t, err)
	require.Empty(t, res)
}

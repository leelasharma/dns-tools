package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDNSNameToMZName(t *testing.T) {
	assert.Equal(t, "example", dnsNameToMZName("example"))
	assert.Equal(t, "example", dnsNameToMZName("example."))
	assert.Equal(t, "com--example", dnsNameToMZName("example.com"))
	assert.Equal(t, "com--example--foo", dnsNameToMZName("foo.example.com"))
	assert.Equal(t, "com--foo-example", dnsNameToMZName("foo-example.com"))
	assert.Equal(t, "uk--co--xn--example", dnsNameToMZName("xn--example.co.uk"))
}

package main

import (
	"strings"
	"testing"
)

func TestGetConstantValues(t *testing.T) {
	testCases := []struct {
		code       string
		valueCount int
	}{
		{
			code:       validCode,
			valueCount: 3,
		},
		{
			code:       invalidCode,
			valueCount: 3,
		},
	}

	for i, tc := range testCases {
		c := newChecker()
		err := c.parse(strings.NewReader(tc.code))
		if err != nil {
			t.Fatal(err)
		}
		err = c.populateConstValues()
		if err != nil {
			t.Fatal(err)
		}
		if len(c.constValues) != 1 {
			t.Fatalf("Expected 1 const type, got %d (Case %d)", len(c.constValues), i)
		}
		for _, v := range c.constValues {
			if len(v) != tc.valueCount {
				t.Fatalf("Expected %d values, got %d (Case %d)", tc.valueCount, len(v), i)
			}
		}
	}
}

func TestAllExhaustive(t *testing.T) {
	testCases := []struct {
		code       string
		exhaustive bool
	}{
		{
			code:       validCode,
			exhaustive: true,
		},
		{
			code:       validCodeWithDefault,
			exhaustive: true,
		},

		{
			code:       invalidCode,
			exhaustive: false,
		},
	}

	for i, tc := range testCases {
		c := newChecker()
		err := c.parse(strings.NewReader(tc.code))
		if err != nil {
			t.Fatal(err)
		}
		err = c.populateConstValues()
		if err != nil {
			t.Fatal(err)
		}
		e := c.allExhaustive()
		if e != tc.exhaustive {
			t.Errorf("Expected %v, Got %v (Case %d)", tc.exhaustive, e, i)
			continue
		}
	}

}

const validCode = `
package thing

type Size int32

const (
	Small Size = iota
	Medium
	Large
)

func Good(s Size) {
	switch s {
	case Small:
	case Medium:
	case Large:
	}
}
`

const validCodeWithDefault = `
package thing

type Size int32

const (
	Small Size = iota
	Medium
	Large
)

func Good(s Size) {
	switch s {
	case Small:
	case Medium:
	default:
	}
}
`

const invalidCode = `
package thing

type Size int32

const (
	Small Size = iota
	Medium
	Large
)

func Good(s Size) {
	switch s {
	case Small:
	case Large:
	}
}
`

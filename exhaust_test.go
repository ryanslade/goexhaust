package main

import (
	"strings"
	"testing"
)

func TestGetConstantValues(t *testing.T) {
	testCases := []struct {
		name       string
		code       string
		valueCount int
	}{
		{
			name:       "Valid Code",
			code:       validCode,
			valueCount: 3,
		},
		{
			name:       "Invalid Code",
			code:       invalidCode,
			valueCount: 3,
		},
		{
			name:       "Invalid Code With String",
			code:       invalidCode,
			valueCount: 3,
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
		})
	}
}

func TestAllExhaustive(t *testing.T) {
	testCases := []struct {
		name       string
		code       string
		exhaustive bool
	}{
		{
			name:       "Valid code",
			code:       validCode,
			exhaustive: true,
		},
		{
			name:       "Valid code with default",
			code:       validCodeWithDefault,
			exhaustive: true,
		},
		{
			name:       "Valid code with two types",
			code:       validCodeTwoTypes,
			exhaustive: true,
		},
		{
			name:       "Valid code with non checkable switch",
			code:       validCodeWithNonCheckableSwitch,
			exhaustive: true,
		},
		{
			name:       "Invalid Code",
			code:       invalidCode,
			exhaustive: false,
		},
		{
			name:       "Invalid Code with string",
			code:       invalidCodeWithString,
			exhaustive: false,
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := newChecker()
			err := c.parse(strings.NewReader(tc.code))
			if err != nil {
				t.Fatal(err)
			}
			err = c.populateConstValues()
			if err != nil {
				t.Fatal(err)
			}
			e, _ := c.allExhaustive()
			if e != tc.exhaustive {
				t.Errorf("Expected %v, Got %v (Case %d)", tc.exhaustive, e, i)
			}
		})
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

const validCodeWithNonCheckableSwitch = `
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

	switch "SomethingElse" {
	case "Something":
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

const validCodeTwoTypes = `
package thing

type Size int32

const (
	Small Size = iota
	Medium
	Large
)

type Another string

const (
	A Another = "A"
	B Another = "B"
)

func Good(s Size) {
	switch s {
	case Small:
	case Medium:
	case Large:
	}
}

func AlsoGood(a Another) {
	switch a {
	case A:
	case B:
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

const invalidCodeWithString = `
package thing

type Size string

const (
	Small Size = "Small"
	Medium Size = "Medium"
	Large Size = "Large"
)

func Good(s Size) {
	switch s {
	case Small:
	case Large:
	}
}
`

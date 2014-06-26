package envdecode

import (
	"fmt"
	"math"
	"os"
	"testing"
)

type nested struct {
	String string `env:"TEST_STRING"`
}

type testConfig struct {
	String  string  `env:"TEST_STRING"`
	Int64   int64   `env:"TEST_INT64"`
	Uint16  uint16  `env:"TEST_UINT16"`
	Float64 float64 `env:"TEST_FLOAT64"`
	Bool    bool    `env:"TEST_BOOL"`

	UnsetString string `env:"TEST_UNSET_STRING"`
	UnsetInt64  int64  `env:"TEST_UNSET_INT64"`

	InvalidInt64 int64 `env:"TEST_INVALID_INT64"`

	UnusedField     string
	unexportedField string

	IgnoredPtr *bool `env:"TEST_BOOL"`

	Nested    nested
	NestedPtr *nested

	Default int `env:"TEST_UNSET,asdf=asdf,default=1234"`
}

type testConfigRequired struct {
	Required string `env:"TEST_REQUIRED,required"`
}
type testConfigRequiredDefault struct {
	RequiredDefault string `env:"TEST_REQUIRED_DEFAULT,required,default=test"`
}

func TestDecode(t *testing.T) {
	os.Setenv("TEST_STRING", "foo")
	os.Setenv("TEST_INT64", fmt.Sprintf("%d", -(1<<50)))
	os.Setenv("TEST_UINT16", "60000")
	os.Setenv("TEST_FLOAT64", fmt.Sprintf("%.48f", math.Pi))
	os.Setenv("TEST_BOOL", "true")
	os.Setenv("TEST_INVALID_INT64", "asdf")

	var tc testConfig
	tc.NestedPtr = &nested{}

	err := Decode(&tc)
	if err != nil {
		t.Fatal(err)
	}

	if tc.String != "foo" {
		t.Fatalf(`Expected "foo", got "%s"`, tc.String)
	}

	if tc.Int64 != -(1 << 50) {
		t.Fatalf("Expected %d, got %d", -(1 << 50), tc.Int64)
	}

	if tc.Uint16 != 60000 {
		t.Fatalf("Expected 60000, got %d", tc.Uint16)
	}

	if tc.Float64 != math.Pi {
		t.Fatalf("Expected %.48f, got %.48f", math.Pi, tc.Float64)
	}

	if !tc.Bool {
		t.Fatal("Expected true, got false")
	}

	if tc.UnsetString != "" {
		t.Fatal("Got non-empty string unexpectedly")
	}

	if tc.UnsetInt64 != 0 {
		t.Fatal("Got non-zero int unexpectedly")
	}

	if tc.InvalidInt64 != 0 {
		t.Fatal("Got non-zero int unexpectedly")
	}

	if tc.UnusedField != "" {
		t.Fatal("Expected empty field")
	}

	if tc.unexportedField != "" {
		t.Fatal("Expected empty field")
	}

	if tc.IgnoredPtr != nil {
		t.Fatal("Expected nil pointer")
	}

	if tc.Nested.String != "foo" {
		t.Fatalf(`Expected "foo", got "%s"`, tc.Nested.String)
	}

	if tc.NestedPtr.String != "foo" {
		t.Fatalf(`Expected "foo", got "%s"`, tc.NestedPtr.String)
	}

	if tc.Default != 1234 {
		t.Fatalf("Expected 1234, got %d", tc.Default)
	}

	os.Setenv("TEST_REQUIRED", "required")
	var tcr testConfigRequired

	err = Decode(&tcr)
	if err != nil {
		t.Fatal(err)
	}

	if tcr.Required != "required" {
		t.Fatalf("Expected \"required\", got %s", tcr.Required)
	}

}

func TestDecodeErrors(t *testing.T) {
	var b bool
	err := Decode(&b)
	if err != ErrInvalidTarget {
		t.Fatal("Should have gotten an error decoding into a bool")
	}

	var tc testConfig
	err = Decode(tc)
	if err != ErrInvalidTarget {
		t.Fatal("Should have gotten an error decoding into a non-pointer")
	}

	var tcp *testConfig
	err = Decode(tcp)
	if err != ErrInvalidTarget {
		t.Fatal("Should have gotten an error decoding to a nil pointer")
	}

	var tcr testConfigRequired
	os.Clearenv()
	err = Decode(&tcr)
	if err == nil {
		t.Fatal("An error was expected but recieved:", err)
	}

	missing := false
	FailureFunc = func(err error) {
		missing = true
	}
	MustDecode(&tcr)
	if !missing {
		t.Fatal("The FailureFunc should have been called but it was not")
	}

	var tcrd testConfigRequiredDefault
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	err = Decode(&tcrd)
	t.Fatal("This should not have been reached. A panic should have occured.")
}

func ExampleDecode() {
	type Example struct {
		// A string field, without any default
		String string `env:"EXAMPLE_STRING"`

		// A uint16 field, with a default value of 100
		Uint16 uint16 `env:"EXAMPLE_UINT16,default=100"`
	}

	os.Setenv("EXAMPLE_STRING", "an example!")

	var e Example
	err := Decode(&e)
	if err != nil {
		panic(err)
	}

	// If TEST_STRING is set, e.String will contain its value
	fmt.Println(e.String)

	// If TEST_UINT16 is set, e.Uint16 will contain its value.
	// Otherwise, it will contain the default value, 100.
	fmt.Println(e.Uint16)

	// Output:
	// an example!
	// 100
}

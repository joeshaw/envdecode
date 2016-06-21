package envdecode

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strconv"
	"testing"
	"time"
)

type nested struct {
	String string `env:"TEST_STRING"`
}

type testConfig struct {
	String   string        `env:"TEST_STRING"`
	Int64    int64         `env:"TEST_INT64"`
	Uint16   uint16        `env:"TEST_UINT16"`
	Float64  float64       `env:"TEST_FLOAT64"`
	Bool     bool          `env:"TEST_BOOL"`
	Duration time.Duration `env:"TEST_DURATION"`
	URL      *url.URL      `env:"TEST_URL"`

	StringSlice   []string        `env:"TEST_STRING_SLICE"`
	Int64Slice    []int64         `env:"TEST_INT64_SLICE"`
	Uint16Slice   []uint16        `env:"TEST_UINT16_SLICE"`
	Float64Slice  []float64       `env:"TEST_FLOAT64_SLICE"`
	BoolSlice     []bool          `env:"TEST_BOOL_SLICE"`
	DurationSlice []time.Duration `env:"TEST_DURATION_SLICE"`
	URLSlice      []*url.URL      `env:"TEST_URL_SLICE"`

	UnsetString   string        `env:"TEST_UNSET_STRING"`
	UnsetInt64    int64         `env:"TEST_UNSET_INT64"`
	UnsetDuration time.Duration `env:"TEST_UNSET_DURATION"`
	UnsetURL      *url.URL      `env:"TEST_UNSET_URL"`
	UnsetSlice    []string      `env:"TEST_UNSET_SLICE"`

	InvalidInt64 int64 `env:"TEST_INVALID_INT64"`

	UnusedField     string
	unexportedField string

	IgnoredPtr *bool `env:"TEST_BOOL"`

	Nested    nested
	NestedPtr *nested

	DecoderStruct    decoderStruct  `env:"TEST_DECODER_STRUCT"`
	DecoderStructPtr *decoderStruct `env:"TEST_DECODER_STRUCT_PTR"`

	DecoderString decoderString `env:"TEST_DECODER_STRING"`

	DefaultInt      int           `env:"TEST_UNSET,asdf=asdf,default=1234"`
	DefaultSliceInt []int         `env:"TEST_UNSET,asdf=asdf,default=1;2;3"`
	DefaultDuration time.Duration `env:"TEST_UNSET,asdf=asdf,default=24h"`
	DefaultURL      *url.URL      `env:"TEST_UNSET,default=http://example.com"`
}

type testConfigRequired struct {
	Required string `env:"TEST_REQUIRED,required"`
}

type testConfigRequiredDefault struct {
	RequiredDefault string `env:"TEST_REQUIRED_DEFAULT,required,default=test"`
}

type testNoExportedFields struct {
	aString  string  `env:"TEST_STRING"`
	anInt64  int64   `env:"TEST_INT64"`
	aUint16  uint16  `env:"TEST_UINT16"`
	aFloat64 float64 `env:"TEST_FLOAT64"`
	aBool    bool    `env:"TEST_BOOL"`
}

type testNoTags struct {
	String string
}

type decoderStruct struct {
	String string
}

func (d *decoderStruct) Decode(env string) error {
	return json.Unmarshal([]byte(env), &d)
}

type decoderString string

func (d *decoderString) Decode(env string) error {
	r, l := []rune(env), len(env)

	for i := 0; i < l/2; i++ {
		r[i], r[l-1-i] = r[l-1-i], r[i]
	}

	*d = decoderString(r)
	return nil
}

func TestDecode(t *testing.T) {
	int64Val := int64(-(1 << 50))
	int64AsString := fmt.Sprintf("%d", int64Val)
	piAsString := fmt.Sprintf("%.48f", math.Pi)

	os.Setenv("TEST_STRING", "foo")
	os.Setenv("TEST_INT64", int64AsString)
	os.Setenv("TEST_UINT16", "60000")
	os.Setenv("TEST_FLOAT64", piAsString)
	os.Setenv("TEST_BOOL", "true")
	os.Setenv("TEST_DURATION", "10m")
	os.Setenv("TEST_URL", "https://example.com")
	os.Setenv("TEST_INVALID_INT64", "asdf")
	os.Setenv("TEST_STRING_SLICE", "foo;bar")
	os.Setenv("TEST_INT64_SLICE", int64AsString+";"+int64AsString)
	os.Setenv("TEST_UINT16_SLICE", "60000;50000")
	os.Setenv("TEST_FLOAT64_SLICE", piAsString+";"+piAsString)
	os.Setenv("TEST_BOOL_SLICE", "true; false; true")
	os.Setenv("TEST_DURATION_SLICE", "10m; 20m")
	os.Setenv("TEST_URL_SLICE", "https://example.com")
	os.Setenv("TEST_DECODER_STRUCT", "{\"string\":\"foo\"}")
	os.Setenv("TEST_DECODER_STRUCT_PTR", "{\"string\":\"foo\"}")
	os.Setenv("TEST_DECODER_STRING", "oof")

	var tc testConfig
	tc.NestedPtr = &nested{}
	tc.DecoderStructPtr = &decoderStruct{}

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

	duration, _ := time.ParseDuration("10m")
	if tc.Duration != duration {
		t.Fatalf("Expected %d, got %d", duration, tc.Duration)
	}

	if tc.URL == nil {
		t.Fatalf("Expected https://example.com, got nil")
	} else if tc.URL.String() != "https://example.com" {
		t.Fatalf("Expected https://example.com, got %s", tc.URL.String())
	}

	expectedStringSlice := []string{"foo", "bar"}
	if !reflect.DeepEqual(tc.StringSlice, expectedStringSlice) {
		t.Fatalf("Expected %s, got %s", expectedStringSlice, tc.StringSlice)
	}

	expectedInt64Slice := []int64{int64Val, int64Val}
	if !reflect.DeepEqual(tc.Int64Slice, expectedInt64Slice) {
		t.Fatalf("Expected %s, got %s", expectedInt64Slice, tc.Int64Slice)
	}

	expectedUint16Slice := []uint16{60000, 50000}
	if !reflect.DeepEqual(tc.Uint16Slice, expectedUint16Slice) {
		t.Fatalf("Expected %s, got %s", expectedUint16Slice, tc.Uint16Slice)
	}

	expectedFloat64Slice := []float64{math.Pi, math.Pi}
	if !reflect.DeepEqual(tc.Float64Slice, expectedFloat64Slice) {
		t.Fatalf("Expected %s, got %s", expectedFloat64Slice, tc.Float64Slice)
	}

	expectedBoolSlice := []bool{true, false, true}
	if !reflect.DeepEqual(tc.BoolSlice, expectedBoolSlice) {
		t.Fatalf("Expected %s, got %s", expectedBoolSlice, tc.BoolSlice)
	}

	duration2, _ := time.ParseDuration("20m")
	expectedDurationSlice := []time.Duration{duration, duration2}
	if !reflect.DeepEqual(tc.DurationSlice, expectedDurationSlice) {
		t.Fatalf("Expected %s, got %s", expectedDurationSlice, tc.DurationSlice)
	}

	urlVal, _ := url.Parse("https://example.com")
	expectedUrlSlice := []*url.URL{urlVal}
	if !reflect.DeepEqual(tc.URLSlice, expectedUrlSlice) {
		t.Fatalf("Expected %s, got %s", expectedUrlSlice, tc.URLSlice)
	}

	if tc.UnsetString != "" {
		t.Fatal("Got non-empty string unexpectedly")
	}

	if tc.UnsetInt64 != 0 {
		t.Fatal("Got non-zero int unexpectedly")
	}

	if tc.UnsetDuration != time.Duration(0) {
		t.Fatal("Got non-zero time.Duration unexpectedly")
	}

	if tc.UnsetURL != nil {
		t.Fatal("Got non-zero *url.URL unexpectedly")
	}

	if len(tc.UnsetSlice) > 0 {
		t.Fatal("Got not-empty string slice unexpectedly")
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

	if tc.DefaultInt != 1234 {
		t.Fatalf("Expected 1234, got %d", tc.DefaultInt)
	}

	expectedDefaultSlice := []int{1, 2, 3}
	if !reflect.DeepEqual(tc.DefaultSliceInt, expectedDefaultSlice) {
		t.Fatalf("Expected %d, got %d", expectedDefaultSlice, tc.DefaultSliceInt)
	}

	defaultDuration, _ := time.ParseDuration("24h")
	if tc.DefaultDuration != defaultDuration {
		t.Fatalf("Expected %d, got %d", defaultDuration, tc.DefaultInt)
	}

	if tc.DefaultURL.String() != "http://example.com" {
		t.Fatalf("Expected http://example.com, got %s", tc.DefaultURL.String())
	}

	if tc.DecoderStruct.String != "foo" {
		t.Fatalf("Expected foo, got %s", tc.DecoderStruct.String)
	}

	if tc.DecoderStructPtr.String != "foo" {
		t.Fatalf("Expected foo, got %s", tc.DecoderStructPtr.String)
	}

	if tc.DecoderString != "foo" {
		t.Fatalf("Expected foo, got %s", tc.DecoderString)
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

	_, err = Export(&tcr)
	if err != nil {
		t.Fatal(err)
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

	var tnt testNoTags
	err = Decode(&tnt)
	if err != ErrInvalidTarget {
		t.Fatal("Should have gotten an error decoding a struct with no tags")
	}

	var tcni testNoExportedFields
	err = Decode(&tcni)
	if err != ErrInvalidTarget {
		t.Fatal("Should have gotten an error decoding a struct with no unexported fields")
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

func TestOnlyNested(t *testing.T) {
	os.Setenv("TEST_STRING", "foo")

	// No env vars in the outer level are ok, as long as they're
	// in the inner struct.
	var o struct {
		Inner nested
	}
	if err := Decode(&o); err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	// No env vars in the inner levels are ok, as long as they're
	// in the outer struct.
	var o2 struct {
		Inner noConfig
		X     string `env:"TEST_STRING"`
	}
	if err := Decode(&o2); err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	// No env vars in either outer or inner levels should result
	// in error
	var o3 struct {
		Inner noConfig
	}
	if err := Decode(&o3); err != ErrInvalidTarget {
		t.Fatal("Expected ErrInvalidTarget, got %s", err)
	}
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

//// Export tests

type testConfigExport struct {
	String   string        `env:"TEST_STRING"`
	Int64    int64         `env:"TEST_INT64"`
	Uint16   uint16        `env:"TEST_UINT16"`
	Float64  float64       `env:"TEST_FLOAT64"`
	Bool     bool          `env:"TEST_BOOL"`
	Duration time.Duration `env:"TEST_DURATION"`
	URL      *url.URL      `env:"TEST_URL"`

	StringSlice []string `env:"TEST_STRING_SLICE"`

	UnsetString   string        `env:"TEST_UNSET_STRING"`
	UnsetInt64    int64         `env:"TEST_UNSET_INT64"`
	UnsetDuration time.Duration `env:"TEST_UNSET_DURATION"`
	UnsetURL      *url.URL      `env:"TEST_UNSET_URL"`

	UnusedField     string
	unexportedField string

	IgnoredPtr *bool `env:"TEST_IGNORED_POINTER"`

	Nested         nestedConfigExport
	NestedPtr      *nestedConfigExportPointer
	NestedPtrUnset *nestedConfigExportPointer

	NestedTwice nestedTwiceConfig

	NoConfig       noConfig
	NoConfigPtr    *noConfig
	NoConfigPtrSet *noConfig

	RequiredInt int `env:"TEST_REQUIRED_INT,required"`

	DefaultBool     bool          `env:"TEST_DEFAULT_BOOL,default=true"`
	DefaultInt      int           `env:"TEST_DEFAULT_INT,default=1234"`
	DefaultDuration time.Duration `env:"TEST_DEFAULT_DURATION,default=24h"`
	DefaultURL      *url.URL      `env:"TEST_DEFAULT_URL,default=http://example.com"`
	DefaultIntSet   int           `env:"TEST_DEFAULT_INT_SET,default=99"`
	DefaultIntSlice []int         `env:"TEST_DEFAULT_INT_SLICE,default=99;33"`
}

type nestedConfigExport struct {
	String string `env:"TEST_NESTED_STRING"`
}

type nestedConfigExportPointer struct {
	String string `env:"TEST_NESTED_STRING_POINTER"`
}

type noConfig struct {
	Int int
}

type nestedTwiceConfig struct {
	Nested nestedConfigInner
}

type nestedConfigInner struct {
	String string `env:"TEST_NESTED_TWICE_STRING"`
}

func TestExport(t *testing.T) {
	testFloat64 := fmt.Sprintf("%.48f", math.Pi)
	testFloat64Output := strconv.FormatFloat(math.Pi, 'f', -1, 64)
	testInt64 := fmt.Sprintf("%d", -(1 << 50))

	os.Setenv("TEST_STRING", "foo")
	os.Setenv("TEST_INT64", testInt64)
	os.Setenv("TEST_UINT16", "60000")
	os.Setenv("TEST_FLOAT64", testFloat64)
	os.Setenv("TEST_BOOL", "true")
	os.Setenv("TEST_DURATION", "10m")
	os.Setenv("TEST_URL", "https://example.com")
	os.Setenv("TEST_STRING_SLICE", "foo;bar")
	os.Setenv("TEST_NESTED_STRING", "nest_foo")
	os.Setenv("TEST_NESTED_STRING_POINTER", "nest_foo_ptr")
	os.Setenv("TEST_NESTED_TWICE_STRING", "nest_twice_foo")
	os.Setenv("TEST_REQUIRED_INT", "101")
	os.Setenv("TEST_DEFAULT_INT_SET", "102")
	os.Setenv("TEST_DEFAULT_INT_SLICE", "1;2;3")

	var tc testConfigExport
	tc.NestedPtr = &nestedConfigExportPointer{}
	tc.NoConfigPtrSet = &noConfig{}

	err := Decode(&tc)
	if err != nil {
		t.Fatal(err)
	}

	rc, err := Export(&tc)
	if err != nil {
		t.Fatal(err)
	}

	expected := []*ConfigInfo{
		&ConfigInfo{
			Field:   "String",
			EnvVar:  "TEST_STRING",
			Value:   "foo",
			UsesEnv: true,
		},
		&ConfigInfo{
			Field:   "Int64",
			EnvVar:  "TEST_INT64",
			Value:   testInt64,
			UsesEnv: true,
		},
		&ConfigInfo{
			Field:   "Uint16",
			EnvVar:  "TEST_UINT16",
			Value:   "60000",
			UsesEnv: true,
		},
		&ConfigInfo{
			Field:   "Float64",
			EnvVar:  "TEST_FLOAT64",
			Value:   testFloat64Output,
			UsesEnv: true,
		},
		&ConfigInfo{
			Field:   "Bool",
			EnvVar:  "TEST_BOOL",
			Value:   "true",
			UsesEnv: true,
		},
		&ConfigInfo{
			Field:   "Duration",
			EnvVar:  "TEST_DURATION",
			Value:   "10m0s",
			UsesEnv: true,
		},
		&ConfigInfo{
			Field:   "URL",
			EnvVar:  "TEST_URL",
			Value:   "https://example.com",
			UsesEnv: true,
		},
		&ConfigInfo{
			Field:   "StringSlice",
			EnvVar:  "TEST_STRING_SLICE",
			Value:   "[foo bar]",
			UsesEnv: true,
		},

		&ConfigInfo{
			Field:  "UnsetString",
			EnvVar: "TEST_UNSET_STRING",
			Value:  "",
		},
		&ConfigInfo{
			Field:  "UnsetInt64",
			EnvVar: "TEST_UNSET_INT64",
			Value:  "0",
		},
		&ConfigInfo{
			Field:  "UnsetDuration",
			EnvVar: "TEST_UNSET_DURATION",
			Value:  "0",
		},
		&ConfigInfo{
			Field:  "UnsetURL",
			EnvVar: "TEST_UNSET_URL",
			Value:  "",
		},

		&ConfigInfo{
			Field:  "IgnoredPtr",
			EnvVar: "TEST_IGNORED_POINTER",
			Value:  "",
		},

		&ConfigInfo{
			Field:   "Nested.String",
			EnvVar:  "TEST_NESTED_STRING",
			Value:   "nest_foo",
			UsesEnv: true,
		},
		&ConfigInfo{
			Field:   "NestedPtr.String",
			EnvVar:  "TEST_NESTED_STRING_POINTER",
			Value:   "nest_foo_ptr",
			UsesEnv: true,
		},

		&ConfigInfo{
			Field:   "NestedTwice.Nested.String",
			EnvVar:  "TEST_NESTED_TWICE_STRING",
			Value:   "nest_twice_foo",
			UsesEnv: true,
		},

		&ConfigInfo{
			Field:    "RequiredInt",
			EnvVar:   "TEST_REQUIRED_INT",
			Value:    "101",
			UsesEnv:  true,
			Required: true,
		},

		&ConfigInfo{
			Field:        "DefaultBool",
			EnvVar:       "TEST_DEFAULT_BOOL",
			Value:        "true",
			DefaultValue: "true",
			HasDefault:   true,
		},
		&ConfigInfo{
			Field:        "DefaultInt",
			EnvVar:       "TEST_DEFAULT_INT",
			Value:        "1234",
			DefaultValue: "1234",
			HasDefault:   true,
		},
		&ConfigInfo{
			Field:        "DefaultDuration",
			EnvVar:       "TEST_DEFAULT_DURATION",
			Value:        "24h0m0s",
			DefaultValue: "24h",
			HasDefault:   true,
		},
		&ConfigInfo{
			Field:        "DefaultURL",
			EnvVar:       "TEST_DEFAULT_URL",
			Value:        "http://example.com",
			DefaultValue: "http://example.com",
			HasDefault:   true,
		},
		&ConfigInfo{
			Field:        "DefaultIntSet",
			EnvVar:       "TEST_DEFAULT_INT_SET",
			Value:        "102",
			DefaultValue: "99",
			HasDefault:   true,
			UsesEnv:      true,
		},
		&ConfigInfo{
			Field:        "DefaultIntSlice",
			EnvVar:       "TEST_DEFAULT_INT_SLICE",
			Value:        "[1 2 3]",
			DefaultValue: "99;33",
			HasDefault:   true,
			UsesEnv:      true,
		},
	}

	sort.Sort(ConfigInfoSlice(expected))

	if len(rc) != len(expected) {
		t.Fatalf("Have %d results, expected %d", len(rc), len(expected))
	}

	for n, v := range rc {
		ci := expected[n]
		if *ci != *v {
			t.Fatalf("have %+v, expected %+v", v, ci)
		}
	}
}

func TestDecodeCustomGetenv(t *testing.T) {
	oldFunc := GetenvFunc
	defer func() {
		GetenvFunc = oldFunc
	}()
	GetenvFunc = func(k string) string {
		if k == "TEST_STRING" {
			return "testing"
		}
		return ""
	}

	var n nested
	err := Decode(&n)

	if err != nil {
		t.Fatal(err)
	}

	if n.String != "testing" {
		t.Fatalf("Have %s expected 'testing'", n.String)
	}
}

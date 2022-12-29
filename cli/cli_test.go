package cli

import (
	"os"
	"reflect"
	"testing"
)

func TestSimpleOneService(t *testing.T) {
	args := []string{
		"--start",
		"FOO",
	}
	result, err := Parse(args)
	if err != nil {
		t.Errorf("parse failed %s", err)
	}

	if result.Start != true {
		t.Error("Expected start to be true")
	}

	if len(result.ExtraServices) != 1 && result.ExtraServices[0] != "FOO" {
		t.Errorf("Expected ExtraServices to have FOO but instead it has %d items", len(result.ExtraServices))
	}
}

func TestComplexOneService(t *testing.T) {
	args := []string{
		"-v",
		"--start",
		"FOO",
		"-r",
		"1.4.33",
	}

	result, err := Parse(args)
	if err != nil {
		t.Errorf("parse failed %s", err)
	}

	if result.Start != true {
		t.Error("Expected start to be true")
	}

	if result.Verbose != true {
		t.Error("Expected verbose to be true")
	}

	if len(result.ExtraServices) != 1 && result.ExtraServices[0] != "FOO" {
		t.Errorf("Expected ExtraServices to have FOO but instead it has %d items", len(result.ExtraServices))
	}

	if result.Release != "1.4.33" {
		t.Errorf("Expected Release to be 1.4.33, instead its %s", result.Release)
	}
}

func TestKitchenSinkOneService(t *testing.T) {
	args := []string{
		"-v",
		"--start",
		"--noprogress",
		"FOO",
		"-r",
		"1.4.33",
		"--clean",
	}

	result, err := Parse(args)
	if err != nil {
		t.Errorf("parse failed %s", err)
	}

	if result.Start != true {
		t.Error("Expected start to be true")
	}

	if result.Verbose != true {
		t.Error("Expected verbose to be true")
	}

	if result.NoProgress != true {
		t.Error("Expected noprogress to be true")
	}

	if result.Clean != true {
		t.Error("Expected clean to be true")
	}

	if len(result.ExtraServices) != 1 && result.ExtraServices[0] != "FOO" {
		t.Errorf("Expected ExtraServices to have FOO but instead it has %d items", len(result.ExtraServices))
	}

	if result.Release != "1.4.33" {
		t.Errorf("Expected Release to be 1.4.33, instead its %s", result.Release)
	}
}

func TestSimpleManyService(t *testing.T) {
	args := []string{
		"--start",
		"FOO",
		"BAR",
		"BAZ",
		"BOP",
		"BIN",
	}

	result, err := Parse(args)
	if err != nil {
		t.Errorf("parse failed %s", err)
	}

	if result.Start != true {
		t.Error("Expected start to be true")
	}

	if len(result.ExtraServices) != 5 {
		t.Errorf("Expected ExtraServices to have 5 items but instead it has %d items", len(result.ExtraServices))
	}

	for i, name := range args[1:] {
		if result.ExtraServices[i] != name {
			t.Errorf("Expected service %s, got %s", name, result.ExtraServices[i])
		}
	}
}

func TestDedupeServices(t *testing.T) {
	args := []string{
		"--start",
		"FOO",
		"FOO",
		"BAZ",
		"FOO",
		"BAZ",
	}

	result, err := Parse(args)
	if err != nil {
		t.Errorf("parse failed %s", err)
	}

	if result.Start != true {
		t.Error("Expected start to be true")
	}

	if len(result.ExtraServices) != 2 {
		t.Errorf("Expected ExtraServices to have 2 items but instead it has %d items", len(result.ExtraServices))
	}

	if result.ExtraServices[0] != "FOO" {
		t.Error("first services wasnt FOO")
	}

	if result.ExtraServices[1] != "BAZ" {
		t.Error("second services wasnt BAZ")
	}

}

func TestComplexManyService(t *testing.T) {
	args := []string{
		"-v",
		"--start",
		"FOO",
		"BAR",
		"BAZ",
		"BOP",
		"BIN",
		"--clean",
		"--noprogress",
	}

	result, err := Parse(args)
	if err != nil {
		t.Errorf("parse failed %s", err)
	}

	if result.Start != true {
		t.Error("Expected start to be true")
	}

	if result.Verbose != true {
		t.Error("Expected start to be true")
	}

	if len(result.ExtraServices) != 5 {
		t.Errorf("Expected ExtraServices to have FOO but instead it has %d items", len(result.ExtraServices))
	}

	for i, name := range args[2:7] {
		if result.ExtraServices[i] != name {
			t.Errorf("Expected service %s, got %s", name, result.ExtraServices[i])
		}
	}
}

func TestServiceWithExtraArgs(t *testing.T) {
	args := []string{
		"--start",
		"FOO",
		"--appendArgs",
		"{\"FOO\": [\"-Dbaz=bar\"]}",
	}

	result, err := Parse(args)
	if err != nil {
		t.Errorf("parse failed %s", err)
	}

	if result.Start != true {
		t.Error("Expected start to be true")
	}

	if len(result.ExtraServices) != 1 && result.ExtraServices[0] != "FOO" {
		t.Errorf("Expected ExtraServices to have FOO but instead it has %d items", len(result.ExtraServices))
		return
	}

	args, ok := result.ExtraArgs["FOO"]
	if !ok {
		t.Errorf("There were no extra args for FOO")
		return
	}
	if args[0] != "-Dbaz=bar" {
		t.Errorf("incorrect extra args for FOO")
	}
}

func TestArgParsing(t *testing.T) {

	json := "{\"SERVICE_ONE\":[\"-DFoo=Bar\",\"SOMETHING\"],\"SERVICE_TWO\":[\"APPEND_THIS\"]}"

	args, err := parseAppendArgs(json)

	if err != nil {
		t.Errorf("Failed to parse extra args: %s", err)
	}

	// check first service
	if s1, ok := args["SERVICE_ONE"]; ok {
		if s1[0] != "-DFoo=Bar" {
			t.Errorf("SERVICE_ONE had incorrect arg in position 0")
		}
		if s1[1] != "SOMETHING" {
			t.Errorf("SERVICE_ONE had incorrect arg in position 1")
		}
	} else {
		t.Errorf("Args did not container SERVICE_ONE")
	}

	// check second service
	if s1, ok := args["SERVICE_TWO"]; ok {
		if s1[0] != "APPEND_THIS" {
			t.Errorf("SERVICE_TWO had incorrect arg in position 0")
		}
	} else {
		t.Errorf("Args did not container SERVICE_TWO")
	}

}

func TestArgParsingFailsOnInvalid(t *testing.T) {

	_, err := parseAppendArgs("-Dhttp.port=123")

	if err == nil {
		t.Error("expected parser to fail, but it didnt")
	}
}

func TestWorkersConfigPrecedence(t *testing.T) {
	args := []string{
		"--start",
		"FOO",
	}

	opts, _ := Parse(args)
	if opts.Workers != 2 {
		t.Error("expected Workers to be 2 when not overridden by environment or args")
	}

	os.Setenv("SM_WORKERS", "5")

	opts, _ = Parse(args)
	if opts.Workers != 5 {
		t.Error("expected Workers to be 5 when not overridden by args")
	}

	args = append(args, "--workers", "10")
	opts, _ = Parse(args)
	if opts.Workers != 10 {
		t.Error("expected Workers to be 10 when supplied in args")
	}

	os.Unsetenv("SM_WORKERS")
}

func TestReleaseIsValid(t *testing.T) {

	validReleases := []string{
		"0.1.2",
		"9999.9999.9999",
		"66.3.0-abcde",
		"55.33.22",
	}

	invalidReleases := []string{
		"",
		"--appendArgs",
		"-f",
		"sometext",
		"{\"foo\":\"bar\"}",
		"55",
	}

	for _, r := range validReleases {
		if releaseIsValid(r) == false {
			t.Errorf("expected %s to be valid, it was not", r)
		}
	}

	for _, r := range invalidReleases {
		if releaseIsValid(r) == true {
			t.Errorf("expected %s to be invalid, it was not", r)
		}
	}
}

func TestRemovalOfMinusRFlag(t *testing.T) {

	tests := map[string]struct {
		in  []string
		out []string
	}{
		"-r at end": {
			in:  []string{"-foo", "-r"},
			out: []string{"-foo"},
		},
		"empty args": {
			in:  []string{},
			out: []string{},
		},
		"-r in the middle": {
			in:  []string{"-start", "FOO", "-r", "-offline"},
			out: []string{"-start", "FOO", "-offline"},
		},
		"-r with a valid version": {
			in:  []string{"-start", "-r", "0.1.2", "FOO"},
			out: []string{"-start", "-r", "0.1.2", "FOO"},
		},
		"-r with service name": {
			in:  []string{"--start", "-r", "FOO"},
			out: []string{"--start", "FOO"},
		},
		"no -r flag": {
			in:  []string{"-start", "FOO:123", "-src"},
			out: []string{"-start", "FOO:123", "-src"},
		},
	}

	for k, test := range tests {
		res := fixupInvalidFlags(test.in)

		if !reflect.DeepEqual(res, test.out) {
			t.Errorf("fixupInvalidFlags: %s failed, %v != %v", k, test.out, res)
		}
	}

}

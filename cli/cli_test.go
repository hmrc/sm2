package cli

import "testing"

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

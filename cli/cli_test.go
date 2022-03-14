package cli

import (
	"testing"
)

func TestSimpleOneService(t *testing.T) {
	args := []string{
		"--start",
		"FOO",
	}
	result := Parse(args)

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
	result := Parse(args)

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
		"--no-progress",
		"FOO",
		"-r",
		"1.4.33",
		"--clean",
	}

	result := Parse(args)

	if result.Start != true {
		t.Error("Expected start to be true")
	}

	if result.Verbose != true {
		t.Error("Expected verbose to be true")
	}

	if result.NoProgress != true {
		t.Error("Expected no-progress to be true")
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
	result := Parse(args)

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
	result := Parse(args)

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
		"--no-progress",
	}
	result := Parse(args)

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

package servicemanager

import (
	"testing"
)

func TestParseServiceAndVersion(t *testing.T) {

	serviceAndVersion := parseServiceAndVersion("CATALOGUE_FRONTEND")
	expectedServiceAndVersion := ServiceAndVersion{"CATALOGUE_FRONTEND", "", ""}
	if serviceAndVersion != expectedServiceAndVersion {
		t.Errorf("Parsed: %#v did not match expected: %#v", serviceAndVersion, expectedServiceAndVersion)
	}

	serviceAndVersion = parseServiceAndVersion("CATALOGUE_FRONTEND_2.11")
	expectedServiceAndVersion = ServiceAndVersion{"CATALOGUE_FRONTEND", "", "2.11"}
	if serviceAndVersion != expectedServiceAndVersion {
		t.Errorf("Parsed: %#v did not match expected: %#v", serviceAndVersion, expectedServiceAndVersion)
	}

	serviceAndVersion = parseServiceAndVersion("CATALOGUE_FRONTEND_2.12")
	expectedServiceAndVersion = ServiceAndVersion{"CATALOGUE_FRONTEND", "", "2.12"}
	if serviceAndVersion != expectedServiceAndVersion {
		t.Errorf("Parsed: %#v did not match expected: %#v", serviceAndVersion, expectedServiceAndVersion)
	}

	serviceAndVersion = parseServiceAndVersion("CATALOGUE_FRONTEND_2.13")
	expectedServiceAndVersion = ServiceAndVersion{"CATALOGUE_FRONTEND", "", "2.13"}
	if serviceAndVersion != expectedServiceAndVersion {
		t.Errorf("Parsed: %#v did not match expected: %#v", serviceAndVersion, expectedServiceAndVersion)
	}

	serviceAndVersion = parseServiceAndVersion("CATALOGUE_FRONTEND_3")
	expectedServiceAndVersion = ServiceAndVersion{"CATALOGUE_FRONTEND", "", "3"}
	if serviceAndVersion != expectedServiceAndVersion {
		t.Errorf("Parsed: %#v did not match expected: %#v", serviceAndVersion, expectedServiceAndVersion)
	}

	serviceAndVersion = parseServiceAndVersion("CATALOGUE_FRONTEND:0.499.0")
	expectedServiceAndVersion = ServiceAndVersion{"CATALOGUE_FRONTEND", "0.499.0", ""}
	if serviceAndVersion != expectedServiceAndVersion {
		t.Errorf("Parsed: %#v did not match expected: %#v", serviceAndVersion, expectedServiceAndVersion)
	}

	serviceAndVersion = parseServiceAndVersion("CATALOGUE_FRONTEND_2.11:0.499.0")
	expectedServiceAndVersion = ServiceAndVersion{"CATALOGUE_FRONTEND", "0.499.0", "2.11"}
	if serviceAndVersion != expectedServiceAndVersion {
		t.Errorf("Parsed: %#v did not match expected: %#v", serviceAndVersion, expectedServiceAndVersion)
	}

	serviceAndVersion = parseServiceAndVersion("CATALOGUE_FRONTEND_2.12:0.499.0")
	expectedServiceAndVersion = ServiceAndVersion{"CATALOGUE_FRONTEND", "0.499.0", "2.12"}
	if serviceAndVersion != expectedServiceAndVersion {
		t.Errorf("Parsed: %#v did not match expected: %#v", serviceAndVersion, expectedServiceAndVersion)
	}

	serviceAndVersion = parseServiceAndVersion("CATALOGUE_FRONTEND_2.13:0.499.0")
	expectedServiceAndVersion = ServiceAndVersion{"CATALOGUE_FRONTEND", "0.499.0", "2.13"}
	if serviceAndVersion != expectedServiceAndVersion {
		t.Errorf("Parsed: %#v did not match expected: %#v", serviceAndVersion, expectedServiceAndVersion)
	}

	serviceAndVersion = parseServiceAndVersion("CATALOGUE_FRONTEND_3:0.499.0")
	expectedServiceAndVersion = ServiceAndVersion{"CATALOGUE_FRONTEND", "0.499.0", "3"}
	if serviceAndVersion != expectedServiceAndVersion {
		t.Errorf("Parsed: %#v did not match expected: %#v", serviceAndVersion, expectedServiceAndVersion)
	}

	serviceAndVersion = parseServiceAndVersion("CATALOGUE_FRONTEND_3:10.11")
	expectedServiceAndVersion = ServiceAndVersion{"CATALOGUE_FRONTEND", "10.11", "3"}
	if serviceAndVersion != expectedServiceAndVersion {
		t.Errorf("Parsed: %#v did not match expected: %#v", serviceAndVersion, expectedServiceAndVersion)
	}
}

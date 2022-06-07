package servicemanager

import (
	"os"
	"testing"
)

func TestJavaVersionCmd(t *testing.T) {
	os.Unsetenv("JAVA_HOME")
	java := javaPath()
	if java != "java" {
		t.Errorf("`java` should be invoked when `$JAVA_HOME` is not defined. Actual: %s", java)
	}

	os.Setenv("JAVA_HOME", "/some/java/location")
	java = javaPath()
	if java != "/some/java/location/bin/java" {
		t.Errorf("`$JAVA_HOME/bin/java` should be invoked when `$JAVA_HOME` is defined. Actual: %s", java)
	}
}

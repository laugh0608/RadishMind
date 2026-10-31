package developmentsandboxrunner

import (
	"bytes"
	"os"
	"testing"
)

func TestRefusesHostExecution(t *testing.T) {
	if os.Getpid() == 1 {
		t.Skip("host refusal test requires a non-init test process")
	}
	var output bytes.Buffer
	if code := Run(1, &output); code != 125 || output.Len() != 0 {
		t.Fatal("host execution accepted", code)
	}
}
func TestOutputIsBoundedWithoutInventingFullCapture(t *testing.T) {
	var output bytes.Buffer
	writer := &limitedOutput{writer: &output, remaining: 3}
	if n, err := writer.Write([]byte("abcdef")); err != nil || n != 6 || !writer.exceeded || output.String() != "abc" {
		t.Fatal(n, err, writer.exceeded, output.String())
	}
	if n, err := writer.Write([]byte("more")); err != nil || n != 4 || output.Len() != 3 {
		t.Fatal(n, err, output.String())
	}
}

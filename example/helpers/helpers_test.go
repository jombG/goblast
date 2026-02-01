package helpers

import "testing"

func TestAdd(t *testing.T) {
	result := Add("Hello", "World")
	expected := "HelloWorld"
	if result != expected {
		t.Errorf("Add(\"Hello\", \"World\") = %q; want %q", result, expected)
	}
}

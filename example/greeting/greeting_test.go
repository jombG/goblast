package greeting

import "testing"

func TestGreet(t *testing.T) {
	result := Greet("World")
	expected := "Hello, World!"
	if result != expected {
		t.Errorf("Greet(\"World\") = %q; want %q", result, expected)
	}
}

package auth

import (
	"reflect"
	"testing"
)

func TestHashPassword(t *testing.T) {
	type test struct {
		input string
		password string
		expected bool
	}

	tests := map[string]test {
		"Correct password": {
			input: "HelloWorld!",
			password: "HelloWorld!",
			expected: true,
		},
		"Incorrect password": {
			input: "HelloWorld!",
			password: "Hello!",
			expected: false,
		},
	}

	for k, v := range tests {
		passwordHash, err := HashPassword(v.password)
		if err != nil {
			t.Fatal(err)
		}

		got, err := CheckPasswordHash(v.input, passwordHash)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(v.expected, got) {
			t.Fatalf("%s: expected: %v, got: %v", k, v.expected, got)
		}
	}
}
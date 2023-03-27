package service

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func BenchmarkSimpleFizzBuzz1000(b *testing.B) {
	s := NewFizzBuzzService(nil)
	for i := 0; i < b.N; i++ {
		s.SimpleFizzBuzz(1000, 3, 5, "fizz", "buzz")
	}
}

func BenchmarkSimpleFizzBuzz10000(b *testing.B) {
	s := NewFizzBuzzService(nil)
	for i := 0; i < b.N; i++ {
		s.SimpleFizzBuzz(10000, 3, 5, "fizz", "buzz")
	}
}

func BenchmarkSimpleFizzBuzz100000(b *testing.B) {
	s := NewFizzBuzzService(nil)
	for i := 0; i < b.N; i++ {
		s.SimpleFizzBuzz(100000, 3, 5, "fizz", "buzz")
	}
}

func BenchmarkSimpleFizzBuzz1000000(b *testing.B) {
	s := NewFizzBuzzService(nil)
	for i := 0; i < b.N; i++ {
		s.SimpleFizzBuzz(1000000, 3, 5, "fizz", "buzz")
	}
}

func TestSimpleFizzBuzz(t *testing.T) {
	tests := []struct {
		name     string
		limit    int
		mod1     int
		mod2     int
		r1       string
		r2       string
		expected []string
	}{
		{
			name:  "basic",
			limit: 15,
			mod1:  3,
			mod2:  5,
			r1:    "fizz",
			r2:    "buzz",
			expected: []string{
				"1",
				"2",
				"fizz",
				"4",
				"buzz",
				"fizz",
				"7",
				"8",
				"fizz",
				"buzz",
				"11",
				"fizz",
				"13",
				"14",
				"fizzbuzz",
			},
		},
		{
			name:  "same modulo",
			limit: 3,
			mod1:  1,
			mod2:  1,
			r1:    "fizz",
			r2:    "buzz",
			expected: []string{
				"fizzbuzz",
				"fizzbuzz",
				"fizzbuzz",
			},
		},
		{
			name:  "same string",
			limit: 3,
			mod1:  1,
			mod2:  1,
			r1:    "test",
			r2:    "test",
			expected: []string{
				"testtest",
				"testtest",
				"testtest",
			},
		},
		{
			name:  "even numbers",
			limit: 4,
			mod1:  2,
			mod2:  4,
			r1:    "two",
			r2:    "four",
			expected: []string{
				"1",
				"two",
				"3",
				"twofour",
			},
		},
	}
	t.Parallel()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fbs := NewFizzBuzzService(nil)
			assert.Equal(t,
				test.expected,
				fbs.SimpleFizzBuzz(test.limit, test.mod1, test.mod2, test.r1, test.r2),
			)
		})
	}
}

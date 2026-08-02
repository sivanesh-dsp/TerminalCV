package shell

import (
	"reflect"
	"testing"
)

func TestTokenize(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"empty", "", nil},
		{"whitespace only", "   \t  ", nil},
		{"single word", "help", []string{"help"}},
		{"collapses spaces", "cat   about.txt", []string{"cat", "about.txt"}},
		{"leading/trailing spaces", "  skills  ", []string{"skills"}},
		{"tabs", "search\tkubernetes", []string{"search", "kubernetes"}},
		{"multi args", "search kubernetes cluster", []string{"search", "kubernetes", "cluster"}},
		{"double quotes group", `search "site reliability"`, []string{"search", "site reliability"}},
		{"single quotes group", `echo 'hello world'`, []string{"echo", "hello world"}},
		{"single quotes keep backslash", `echo 'a\b'`, []string{"echo", `a\b`}},
		{"backslash escape space", `search foo\ bar`, []string{"search", "foo bar"}},
		{"escaped quote", `echo \"hi\"`, []string{"echo", `"hi"`}},
		{"empty quoted arg", `echo ""`, []string{"echo", ""}},
		{"adjacent quotes concat", `echo a"b"c`, []string{"echo", "abc"}},
		{"sudo hire-me", "sudo hire-me", []string{"sudo", "hire-me"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Tokenize(tc.in)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("Tokenize(%q) = %#v, want %#v", tc.in, got, tc.want)
			}
		})
	}
}

package main

import "testing"

func TestVerifyLoginDecision(t *testing.T) {
	tests := []struct {
		name   string
		matter Matter
		err    error
		want   bool
	}{
		{"verified matter", Matter{ID: "M-17", Phone: "+15550001111"}, nil, true},
		{"provider rejection", Matter{ID: "M-17", Phone: "+15550001111"}, errSentinel{}, false},
		{"missing matter", Matter{Phone: "+15550001111"}, nil, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := verifyLogin(tc.matter, func(string, string) error { return tc.err }, "123456")
			if d.Allowed != tc.want {
				t.Fatalf("allowed=%v want %v", d.Allowed, tc.want)
			}
		})
	}
}

type errSentinel struct{}

func (errSentinel) Error() string { return "invalid code" }

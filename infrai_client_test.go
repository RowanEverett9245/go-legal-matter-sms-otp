package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestSMSRequestFields(t *testing.T) {
	tests := []struct {
		name, path string
		want       map[string]string
		call       func(*SMSClient) error
	}{
		{"otp", "/v1/sms/otp", map[string]string{"to": "+15550001111"}, func(c *SMSClient) error { return c.OTP("+15550001111") }},
		{"verify", "/v1/sms/verify", map[string]string{"to": "+15550001111", "code": "123456"}, func(c *SMSClient) error { return c.Verify("+15550001111", "123456") }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != tc.path {
					t.Errorf("request = %s %s, want POST %s", r.Method, r.URL.Path, tc.path)
				}
				var got map[string]string
				if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
					t.Errorf("decode request: %v", err)
				}
				if !reflect.DeepEqual(got, tc.want) {
					t.Errorf("body = %v, want %v", got, tc.want)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"ok":true,"data":{}}`))
			}))
			defer server.Close()
			client := &SMSClient{base: server.URL, key: "test", http: server.Client()}
			if err := tc.call(client); err != nil {
				t.Fatal(err)
			}
		})
	}
}

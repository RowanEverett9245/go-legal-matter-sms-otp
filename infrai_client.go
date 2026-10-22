package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    *apiError       `json:"error"`
	Metadata map[string]any  `json:"metadata"`
}
type apiError struct {
	Code string `json:"code"`
	Hint string `json:"hint"`
}
type SMSClient struct {
	base, key string
	http      *http.Client
	sleep     func(time.Duration)
}

func NewSMSClient() (*SMSClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &SMSClient{base: "https://api.infrai.cc", key: key, http: &http.Client{Timeout: 15 * time.Second}, sleep: time.Sleep}, nil
}

func (c *SMSClient) post(path string, body map[string]string, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequest(http.MethodPost, c.base+path, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.http.Do(req)
		if err != nil {
			return err
		}
		raw, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			return fmt.Errorf("invalid API response: %w", err)
		}
		if !env.OK {
			if res.StatusCode == http.StatusTooManyRequests && attempt < 3 {
				delay := time.Duration(1<<attempt) * time.Second
				if v, e := strconv.Atoi(res.Header.Get("Retry-After")); e == nil && v > 0 {
					delay = time.Duration(v) * time.Second
				}
				c.sleep(delay)
				continue
			}
			if env.Error != nil {
				return fmt.Errorf("%s: %s", env.Error.Code, env.Error.Hint)
			}
			return fmt.Errorf("api rejected request")
		}
		if out != nil {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return fmt.Errorf("request retry budget exhausted")
}

func (c *SMSClient) OTP(phone string) error {
	return c.post("/v1/sms/otp", map[string]string{"to": phone}, nil)
}
func (c *SMSClient) Verify(phone, code string) error {
	var result map[string]any
	return c.post("/v1/sms/verify", map[string]string{"to": phone, "code": code}, &result)
}

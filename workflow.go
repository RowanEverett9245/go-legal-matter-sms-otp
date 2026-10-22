package main

import "fmt"

type Matter struct {
	ID, Phone                        string
	SignedDocument, DeadlineFollowUp bool
}
type LoginDecision struct {
	Allowed bool
	Reason  string
}

func verifyLogin(m Matter, verify func(string, string) error, code string) LoginDecision {
	if m.ID == "" || m.Phone == "" {
		return LoginDecision{Reason: "matter and phone are required"}
	}
	if err := verify(m.Phone, code); err != nil {
		return LoginDecision{Reason: fmt.Sprintf("code rejected: %v", err)}
	}
	return LoginDecision{Allowed: true, Reason: "verified legal matter login"}
}

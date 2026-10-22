package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	client, err := NewSMSClient()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/login/request", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var m Matter
		if json.NewDecoder(r.Body).Decode(&m) != nil || m.Phone == "" {
			http.Error(w, "phone is required", http.StatusBadRequest)
			return
		}
		if err := client.OTP(m.Phone); err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		writeJSON(w, map[string]string{"status": "code requested", "matter_id": m.ID})
	})
	http.HandleFunc("/login/verify", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var in struct {
			Matter Matter `json:"matter"`
			Code   string `json:"code"`
		}
		if json.NewDecoder(r.Body).Decode(&in) != nil {
			http.Error(w, "invalid json", 400)
			return
		}
		decision := verifyLogin(in.Matter, client.Verify, in.Code)
		status := http.StatusOK
		if !decision.Allowed {
			status = http.StatusUnauthorized
		}
		writeJSONStatus(w, status, decision)
	})
	log.Println("legal OTP service listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
func writeJSON(w http.ResponseWriter, v any) { writeJSONStatus(w, http.StatusOK, v) }
func writeJSONStatus(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

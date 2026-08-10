package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "baseline"
	}
	paymentsURL := os.Getenv("PAYMENTS_API_URL")
	if paymentsURL == "" {
		paymentsURL = "http://payments-api:8080"
	}
	accountsURL := os.Getenv("ACCOUNTS_API_URL")
	if accountsURL == "" {
		accountsURL = "http://accounts-api:8080"
	}

	client := &http.Client{Timeout: 5 * time.Second}

	proxyRequest := func(w http.ResponseWriter, r *http.Request, target string) {
		req, _ := http.NewRequestWithContext(r.Context(), r.Method, target, r.Body)
		if pid := r.Header.Get("x-preview-id"); pid != "" {
			req.Header.Set("x-preview-id", pid)
		}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "gateway", "version": version})
	})

	http.HandleFunc("/api/payments", func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, paymentsURL+"/api/payments")
	})

	http.HandleFunc("/api/accounts", func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, accountsURL+"/api/accounts/balance")
	})

	log.Printf("gateway %s listening on :%s", version, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

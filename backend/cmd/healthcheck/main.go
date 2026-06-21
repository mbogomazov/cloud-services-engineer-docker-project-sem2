// Command healthcheck is a tiny static probe used by Docker HEALTHCHECK.
// The final backend image is distroless (no shell, no curl/wget), so the
// container health is verified by this compiled binary hitting /health.
package main

import (
	"net/http"
	"os"
	"time"
)

func main() {
	addr := os.Getenv("HEALTHCHECK_URL")
	if addr == "" {
		addr = "http://127.0.0.1:8081/health"
	}

	client := &http.Client{Timeout: 2 * time.Second}

	resp, err := client.Get(addr)
	if err != nil {
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}
}

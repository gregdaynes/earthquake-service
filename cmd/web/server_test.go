package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

func Test(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	endpoint := "http://0.0.0.0:4000/"
	t.Cleanup(cancel)

	go Run(ctx)
	waitForReady(ctx, 1, endpoint)

	res, err := http.Get(endpoint)
	if err != nil {
		t.Errorf("Error making request: %s\n", err.Error())
	}
	defer res.Body.Close()

	out, err := io.ReadAll(res.Body)
	got := string(out)
	want := "Testing"

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	res.Body.Close()
}

// waitForReady calls the specified endpoint until it gets a 200
// response or until the context is cancelled or the timeout is
// reached.
func waitForReady(
	ctx context.Context,
	timeout time.Duration,
	endpoint string,
) error {
	client := http.Client{}
	startTime := time.Now()
	for {
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			endpoint,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		res, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error making request: %s\n", err.Error())
			continue
		}
		if res.StatusCode == http.StatusOK {
			fmt.Println("Endpoint is ready!")
			res.Body.Close()
			return nil
		}
		res.Body.Close()

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if time.Since(startTime) >= timeout {
				return fmt.Errorf("timeout reached while waiting for endpoint")
			}
			// wait a little while between checks
			time.Sleep(250 * time.Millisecond)
		}
	}
}

//go:build docker
// +build docker

package tests

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_Docker_ContainerStartup(t *testing.T) {
	// Check if docker-compose is installed
	_, err := exec.LookPath("docker-compose")
	if err != nil {
		t.Skip("docker-compose not found")
	}

	// 1. Start Container
	// Ensure we are in root (using -f ../docker-compose.yml might be needed if running from tests dir)
	// Actually go test runs in directory. So ../docker-compose.yml
	cmd := exec.Command("docker-compose", "-f", "../docker-compose.yml", "up", "-d")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "docker-compose up failed: %s", out)

	defer func() {
		// Teardown
		exec.Command("docker-compose", "-f", "../docker-compose.yml", "down").Run()
	}()

	// 2. Wait for ports (Poll for 30s)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check Port 25 (SMTP)
	err = waitForPort(ctx, "localhost:25")
	assert.NoError(t, err, "SMTP port 25 not accessible")

	// Check Port 80 (HTTP/ACME)
	err = waitForPort(ctx, "localhost:80")
	assert.NoError(t, err, "HTTP port 80 not accessible")

	// Check Port 443 (HTTPS)
	err = waitForPort(ctx, "localhost:443")
	assert.NoError(t, err, "HTTPS port 443 not accessible")
}

func waitForPort(ctx context.Context, address string) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for %s", address)
		case <-ticker.C:
			conn, err := net.DialTimeout("tcp", address, 1*time.Second)
			if err == nil {
				conn.Close()
				return nil
			}
		}
	}
}

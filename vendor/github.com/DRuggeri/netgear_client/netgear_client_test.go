package netgear_client

import (
	"os"
	"testing"
)

func TestLogin(t *testing.T) {
	debug := false

	if get_password() == "" {
		t.Fatal("Error: NETGEAR_PASSWORD environment variable is not set")
	}

	client, err := NewNetgearClient(get_url(), true, get_username(), get_password(), 2, debug)
	if err != nil {
		t.Fatalf("Error getting a client: %s", err)
	}

	err = client.LogIn()
	if err != nil {
		t.Fatalf("Error logging in: %s", err)
	}
}

func get_username() string {
	if os.Getenv("NETGEAR_USERNAME") != "" {
		return os.Getenv("NETGEAR_USERNAME")
	}
	return "admin"
}
func get_password() string {
	return os.Getenv("NETGEAR_PASSWORD")
}
func get_url() string {
	if os.Getenv("NETGEAR_URL") != "" {
		return os.Getenv("NETGEAR_URL")
	}
	return "https://www.routerlogin.com"
}

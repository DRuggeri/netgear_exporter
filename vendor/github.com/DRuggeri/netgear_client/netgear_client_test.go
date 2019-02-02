package netgear_client

import (
	"os"
	"testing"
)

func TestLogin(t *testing.T) {
	debug := false

	if password() == "" {
		t.Fatal("Error: NETGEAR_PASSWORD environment variable is not set")
	}

	client, err := NewNetgearClient(url(), true, username(), password(), 2, debug)
	if err != nil {
		t.Fatalf("Error getting a client: %s", err)
	}

	err = client.LogIn()
	if err != nil {
		t.Fatalf("Error logging in: %s", err)
	}
}

func TestGetTrafficMeterStatistics(t *testing.T) {
	debug := false

	if password() == "" {
		t.Fatal("Error: NETGEAR_PASSWORD environment variable is not set")
	}

	client, err := NewNetgearClient(url(), true, username(), password(), 2, debug)
	if err != nil {
		t.Fatalf("Error getting a client: %s", err)
	}

	res, err := client.GetTrafficMeterStatistics()
	if err != nil {
		t.Fatalf("Error getting traffic statistics: %s", err)
	}
	if len(res) < 1 {
		t.Fatalf("Result is empty... WTF!")
	}
}

func username() string {
	if os.Getenv("NETGEAR_USERNAME") != "" {
		return os.Getenv("NETGEAR_USERNAME")
	}
	return "admin"
}
func password() string {
	return os.Getenv("NETGEAR_PASSWORD")
}
func url() string {
	if os.Getenv("NETGEAR_URL") != "" {
		return os.Getenv("NETGEAR_URL")
	}
	return "https://www.routerlogin.com"
}

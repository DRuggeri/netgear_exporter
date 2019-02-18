package netgear_client

import (
	"fmt"
	"testing"
)

func TestGetSystemInfo(t *testing.T) {
	debug := true

	if get_password() == "" {
		t.Fatal("Error: NETGEAR_PASSWORD environment variable is not set")
	}

	client, err := NewNetgearClient(get_url(), true, get_username(), get_password(), 2, debug)
	if err != nil {
		t.Fatalf("Error getting a client: %s", err)
	}

	res, err := client.GetSystemInfo()
	if err != nil {
		t.Fatalf("Error getting traffic statistics: %s", err)
	}
	if len(res) < 1 {
		t.Fatalf("Result is empty... WTF!")
	}

	expected := [...]string{
		"PhysicalFlash",
		"AvailableFlash",
		"CPUUtilization",
		"PhysicalMemory",
		"MemoryUtilization",
	}
	for _, key := range expected {
		if _, ok := res[key]; !ok {
			t.Errorf("Expected `%s` key in the response, but did not find it", key)
		}
	}

	if debug {
		for k, v := range res {
			fmt.Printf("%v => %v\n", k, v)
		}
	}
}

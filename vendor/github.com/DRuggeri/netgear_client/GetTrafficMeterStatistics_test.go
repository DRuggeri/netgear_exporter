package netgear_client

import (
	"fmt"
	"testing"
)

func TestGetTrafficMeterStatistics(t *testing.T) {
	debug := true

	if get_password() == "" {
		t.Fatal("Error: NETGEAR_PASSWORD environment variable is not set")
	}

	client, err := NewNetgearClient(get_url(), true, get_username(), get_password(), 2, debug)
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
	if debug {
		for k, v := range res {
			fmt.Printf("%v => %v\n", k, v)
		}
	}
}

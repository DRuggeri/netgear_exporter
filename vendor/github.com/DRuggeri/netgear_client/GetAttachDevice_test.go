package netgear_client

import (
	"fmt"
	"testing"
)

func TestGetAttachDevice(t *testing.T) {
	debug := true

	if get_password() == "" {
		t.Fatal("Error: NETGEAR_PASSWORD environment variable is not set")
	}

	client, err := NewNetgearClient(get_url(), true, get_username(), get_password(), 10, debug)
	if err != nil {
		t.Fatalf("Error getting a client: %s", err)
	}

	res, err := client.GetAttachDevice()
	if err != nil {
		t.Fatalf("Error getting traffic statistics: %s", err)
	}
	if len(res) < 1 {
		t.Fatalf("Result is empty... WTF!")
	}

	if debug {
		for i, data := range res {
			fmt.Printf("Client %d\n", i)
			for k, v := range data {
				fmt.Printf("  %v => %v\n", k, v)
			}
		}
	}

	if _, ok := res[0]["Name"]; !ok {
		t.Fatalf("The first item in the list didn't contain a name. All items should have a Name!")
	}
}

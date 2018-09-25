package main

import (
	"fmt"
	"testing"
)

func TestExec(t *testing.T) {
	c, err := NewConnectItemWithOpts(withNode("192.168.33.11"), withContainer("23deae9fc89d"), withPort(2375))
	if err != nil {
		t.Fatal(err)
	}

	res, err := c.exec()
	if err != nil {
		if sErr, ok := err.(*ServerError); ok {
			t.Fatalf("\nStatusCode: %d\nMessage: %s\n", sErr.StatusCode, sErr.Message)
		}
		t.Fatal(err)
	}

	fmt.Printf("%+v\n", res)
}

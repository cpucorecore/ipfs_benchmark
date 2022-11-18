package main

import "testing"

func TestUrl(t *testing.T) {
	httpInput := HttpInput{
		Host:   "127.0.0.1",
		Port:   "9094",
		Method: "POST",
		Path:   "/api/v0/swarm/peers",
	}
	repeatHttpInput := RepeatHttpInput{HttpInput: httpInput}
	t.Log(repeatHttpInput.url(""))

	httpInput.Path = "/pins/ipfs"
	iterUrlHttpInput := IterUrlHttpInput{HttpInput: httpInput}
	t.Log(iterUrlHttpInput.url("QmPG8j4qPYkJr6wHemHK1u7jCa6hYbu3d53CZa5yP5aomz"))
}

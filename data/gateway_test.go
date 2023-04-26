package data

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

func TestDataReadBuf(t *testing.T) {
	bodyBytes := []byte("This is the body of the request")
	req, err := http.NewRequest(
		"GET",
		"https://foo.com",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Read the buffer
	bytesRead, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(bytesRead) < len(bodyBytes) {
		t.Errorf("Only read %d bytes", len(bytesRead))
	}
	bytesRead, err = io.ReadAll(req.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(bytesRead) < len(bodyBytes) {
		t.Errorf("Only read %d bytes", len(bytesRead))
	}
}

package widevine

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

type device struct {
	dir         string
	private_key string
	client_id   string
}

var Android = device{
	dir:         "../device",
	private_key: "private_key.pem",
	client_id:   "client_id.bin",
}

func TestCtv(t *testing.T) {
	key, err := base64.StdEncoding.DecodeString(ctv_ca.key)
	if err != nil {
		t.Fatal(err)
	}

	privateKeyPath := filepath.Join(Android.dir, Android.private_key)
	clientIDPath := filepath.Join(Android.dir, Android.client_id)

	private_key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		t.Fatal(err)
	}
	client_id, err := os.ReadFile(clientIDPath)
	if err != nil {
		t.Fatal(err)
	}

	key_id, err := base64.StdEncoding.DecodeString(ctv_ca.key_id)
	if err != nil {
		t.Fatal(err)
	}

	var pssh1 Pssh
	pssh1.KeyIds = [][]byte{key_id}
	pssh1.ContentId, err = base64.StdEncoding.DecodeString(ctv_ca.content_id)
	if err != nil {
		t.Fatal(err)
	}

	var cdm1 Cdm
	err = cdm1.New(private_key, client_id, pssh1.Marshal())
	if err != nil {
		t.Fatal(err)
	}

	data, err := cdm1.RequestBody()
	if err != nil {
		t.Fatal(err)
	}

	resp, err := ctv_service(data)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("license server error: %s", resp.Status)
	}

	var body ResponseBody
	err = body.Unmarshal(data)
	if err != nil {
		t.Fatal(err)
	}

	block, err := cdm1.Block(body)
	if err != nil {
		t.Fatal(err)
	}

	for container := range body.Container() {
		if bytes.Equal(container.Id(), key_id) {
			if bytes.Equal(container.Key(block), key) {
				return
			}
		}
	}
	t.Fatal("key not found")
}

func ctv_service(data []byte) (*http.Response, error) {
	return http.Post(
		"https://license.9c9media.ca/widevine", "application/x-protobuf",
		bytes.NewReader(data),
	)
}

var ctv_ca = struct {
	content_id string
	key        string
	key_id     string
	url        string
}{
	content_id: "ZmYtOGYyNjEzYWUtNTIxNTAx",
	key:        "xQ87t+z5cLOVgxDdSgHyoA==",
	key_id:     "A98dtspZsb9/z++3IHp0Dw==",
	url:        "ctv.ca/movies/fools-rush-in-57470",
}

package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/quelcom/homeink/config"
	"github.com/quelcom/homeink/kindle"
)

const testConfig = `
Port = ":3000"
KindleIP = "192.168.2.2"
KindlePassword = ""
PathToFBInk = "/mnt/us/developer/fbink"
`

func mockKindleRenderImage(path string, row int, col int) error {
	return nil
}

func TestWater(t *testing.T) {
	config.EmbedTestConfig(testConfig)

	litersValue := 42
	litersPayload := fmt.Sprintf(`{"liters":%d}`, litersValue)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/water", bytes.NewReader([]byte(litersPayload)))
	w := httptest.NewRecorder()
	handleWater(w, req)
	res := w.Result()
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("got error %v", err)
	}

	wantStatus := 200
	if res.StatusCode != wantStatus {
		t.Errorf("got status %d, want status %d", res.StatusCode, wantStatus)
	}

	messageStr := fmt.Sprintf("Hello World. Liters are %d", litersValue)
	var want bytes.Buffer
	_ = json.NewEncoder(&want).Encode(map[string]string{"message": messageStr})

	if string(data) != want.String() {
		t.Errorf("got %q, want %q", string(data), want.String())
	}
}

func TestHourlyScreenshot(t *testing.T) {
	config.EmbedTestConfig(testConfig)

	kindle.RenderImage = mockKindleRenderImage
	defer func() { kindle.RenderImage = kindle.ActualRenderImage }()

	r, err := os.Open("../testdata/supersaa.png")
	if err != nil {
		t.Fatalf("test file could not be opened. Error: " + err.Error())
	}

	values := map[string]io.Reader{
		"image":    r,
		"filename": strings.NewReader("supersaa-hourly.png"),
	}

	statusCode, response, err := _doMultipartRequest(values, "/api/v1/screenshot")
	if err != nil {
		t.Fatal(err)
	}

	wantStatus := 200
	if statusCode != wantStatus {
		t.Errorf("got status %d, want status %d", statusCode, wantStatus)
	}

	var want bytes.Buffer
	_ = json.NewEncoder(&want).Encode(map[string]string{"message": "Everything ok"})

	if response != want.String() {
		t.Errorf("got %q, want %q", response, want.String())
	}

	if fileChecksum("supersaa.png") != fileChecksum("../testdata/supersaa.png") {
		t.Errorf("checksums do not match")
	}

	err = os.Remove("supersaa.png")
	if err != nil {
		t.Errorf("could not remove test file. Error: %v", err.Error())
	}
}

func TestInvalidScreenshot(t *testing.T) {
	config.EmbedTestConfig(testConfig)

	kindle.RenderImage = mockKindleRenderImage
	defer func() { kindle.RenderImage = kindle.ActualRenderImage }()

	r, err := os.Open("../testdata/supersaa.png")
	if err != nil {
		t.Fatalf("test file could not be opened. Error: " + err.Error())
	}

	values := map[string]io.Reader{
		"image":    r,
		"filename": strings.NewReader("unknown"),
	}

	statusCode, response, err := _doMultipartRequest(values, "/api/v1/screenshot")
	if err != nil {
		t.Fatal(err)
	}

	wantStatus := 400
	if statusCode != wantStatus {
		t.Errorf("got status %d, want status %d", statusCode, wantStatus)
	}

	wantResponse := fmt.Sprintf("Filename %q not accepted\n", "unknown")

	if response != wantResponse {
		t.Errorf("got %q, want %q", response, wantResponse)
	}
}

func _doMultipartRequest(values map[string]io.Reader, url string) (int, string, error) {
	b, multipart, err := _prepareMultipartBuffer(values)
	if err != nil {
		return 0, "", errors.New("could not prepare multipart form. Error: " + err.Error())
	}

	req := httptest.NewRequest(http.MethodGet, url, &b)
	req.Header.Set("Content-Type", multipart.FormDataContentType())

	w := httptest.NewRecorder()
	handleScreenshot(w, req)
	res := w.Result()
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)

	if err != nil {
		return 0, "", errors.New("error reading response body. Error: %v" + err.Error())
	}

	return res.StatusCode, string(data), nil
}

// Adapted from https://stackoverflow.com/a/20397167
func _prepareMultipartBuffer(values map[string]io.Reader) (bytes.Buffer, *multipart.Writer, error) {
	var b bytes.Buffer
	multipart := multipart.NewWriter(&b)
	var err error

	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an image file
		if x, ok := r.(*os.File); ok {
			if fw, err = multipart.CreateFormFile(key, x.Name()); err != nil {
				return b, nil, err
			}
		} else {
			// Add other fields
			if fw, err = multipart.CreateFormField(key); err != nil {
				return b, nil, err
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return b, nil, err
		}
	}

	multipart.Close()
	return b, multipart, nil
}

func fileChecksum(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

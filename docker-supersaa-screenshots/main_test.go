package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"supersaa/config"
	c "supersaa/config"
	"testing"
)

const testHTTPServerPort = ":8000"

var localTestURL string

func TestMain(m *testing.M) {
	localTestURL = "http://" + getLocalIP() + testHTTPServerPort + "/"

	go runTestHTTPServer()
	code := m.Run()
	os.Exit(code)
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			fmt.Println(ipNet.IP)
			if strings.HasPrefix(ipNet.IP.String(), "192") {
				return ipNet.IP.String()
			}
		}
	}

	return ""
}

// docker run -d -p 9222:9222 --rm --name headless-shell chromedp/headless-shell
func TestHourlyScreenshot(t *testing.T) {
	const testConfig = `
UseRemoteAllocator = true
ScreenshotDir = "/tmp/screenshot/"

[[SupersaaScreenshot]]
Filename = "supersaa-hourly.png"
Selection = "//h2[contains(text(), 'Tuntikohtainen ennuste')]/ancestor::table"
Actions = [
	# Remove show more
	"let showMore = Array.from(document.querySelectorAll('span')).find(el => el.textContent === 'NÄYTÄ LISÄÄ').closest('button'); if (showMore) showMore.remove();"
]
`
	if !isHeadlessShellRunning() {
		t.Skip("Headless shell not running")
	}

	config.EmbedTestConfig(testConfig)

	screenshot := c.GetConfig().SupersaaScreenshot[0]

	t.Run("Get screenshot from hourly section", func(t *testing.T) {
		screenshotBytes := getScreenshotFromElement(localTestURL, screenshot)
		err := saveImageToGrayScale(screenshot.Filename, screenshotBytes)
		if err != nil {
			t.Fatal("Could not save image. Err: " + err.Error())
		}
	})

	t.Run("Validate daily screenshot against golden daily screenshot", func(t *testing.T) {
		expectedFileContents, _ := os.ReadFile("./testdata/expected-supersaa-hourly.png")
		expectedFileChecksum := computeShasum(expectedFileContents)

		gotFileContents, _ := os.ReadFile("/tmp/screenshot/" + screenshot.Filename)
		gotFileChecksum := computeShasum(gotFileContents)

		if expectedFileChecksum != gotFileChecksum {
			t.Errorf("Unexpected file checksum. Expected %s but got %s", expectedFileChecksum, gotFileChecksum)
		}
	})
}

func TestWeeklyScreenshot(t *testing.T) {
	const testConfig = `
UseRemoteAllocator = true
ScreenshotDir = "/tmp/screenshot/"

[[SupersaaScreenshot]]
Filename = "supersaa-weekly.png"
Selection = "//h2[contains(text(), 'Viikon sääennuste')]/ancestor::table"
Actions = [
        # Remove dropdown arrow to expand hourly foreacast for the day
        "let dropArrow = document.querySelectorAll('button.show-more-button-text'); dropArrow.forEach(function(el) { el.remove(); })"
]
`

	if !isHeadlessShellRunning() {
		t.Skip("Headless shell not running")
	}

	config.EmbedTestConfig(testConfig)

	screenshot := c.GetConfig().SupersaaScreenshot[0]

	t.Run("Get screenshot from weekly section", func(t *testing.T) {
		screenshotBytes := getScreenshotFromElement(localTestURL, screenshot)
		err := saveImageToGrayScale(screenshot.Filename, screenshotBytes)
		if err != nil {
			t.Fatal("Could not save image. Err: " + err.Error())
		}
	})

	t.Run("Validate weekly screenshot against golden weekly screenshot", func(t *testing.T) {
		expectedFileContents, _ := os.ReadFile("./testdata/expected-supersaa-weekly.png")
		expectedFileChecksum := computeShasum(expectedFileContents)

		gotFileContents, _ := os.ReadFile("/tmp/screenshot/" + screenshot.Filename)
		gotFileChecksum := computeShasum(gotFileContents)

		if expectedFileChecksum != gotFileChecksum {
			t.Errorf("Unexpected file checksum. Expected %s but got %s", expectedFileChecksum, gotFileChecksum)
		}
	})
}

func runTestHTTPServer() {
	wd, err := os.Getwd()
	if err != nil {
		panic("Cannot get working directory")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, wd+"/testdata/supersaa_oulu_2024-07-13_1612.html")
	})

	_ = http.ListenAndServe(":8000", nil)
}

func isHeadlessShellRunning() bool {
	res, err := http.Get("http://localhost:9222/json/version")
	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
		return false
	}

	return res.StatusCode == 200
}

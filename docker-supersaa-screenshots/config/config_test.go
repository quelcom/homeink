package config

import (
	"strings"
	"testing"
)

const testConfig = `
ScreenshotDir = "/tmp/screenshot/"

[[SupersaaScreenshot]]
Filename = "supersaa-hourly.png"
Selection = "//h2[contains(text(), 'Tuntikohtainen ennuste')]/ancestor::table"
Actions = [
	# Remove show more
	"let showMore = Array.from(document.querySelectorAll('span')).find(el => el.textContent === 'NÄYTÄ LISÄÄ').closest('button'); if (showMore) showMore.remove();"
]

[[SupersaaScreenshot]]
Filename = "supersaa-weekly.png"
Selection = "//h2[contains(text(), 'Viikon sääennuste')]/ancestor::table"
Actions = [
	# Remove dropdown arrow to expand hourly foreacast for the day
	"let dropArrow = document.querySelectorAll('button.show-more-button-text'); dropArrow.forEach(function(el) { el.remove(); });"
]
`

func TestEmbedConfig(t *testing.T) {
	EmbedTestConfig(testConfig)
	size := len(GetConfig().SupersaaScreenshot)
	if size != 2 {
		t.Errorf("Expected 2, got %d", size)
	}

	for _, v := range GetConfig().SupersaaScreenshot {
		if !strings.HasPrefix(v.Filename, "supersaa-") {
			t.Errorf("Expected %s string to start with supersaa-", v.Filename)
		}
	}

	hourly := GetConfig().SupersaaScreenshot[0]
	expected := "//h2[contains(text(), 'Tuntikohtainen ennuste')]/ancestor::table"
	if hourly.Selection != expected {
		t.Errorf("Expected %s but got %s", expected, hourly.Selection)
	}

	size = len(hourly.Actions)
	if size != 1 {
		t.Errorf("Expected 1, got %d", size)
	}

	if !strings.HasPrefix(hourly.Actions[0], "let showMore") {
		t.Errorf("Expected %s string to start with let showMore", hourly.Actions[0])
	}

	if !strings.HasSuffix(hourly.Actions[0], "remove();") {
		t.Errorf("Expected %s string to start with remove();", hourly.Actions[0])
	}

}

package main

import (
	"bytes"
	"context"
	"crypto/sha1"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	c "supersaa/config"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// Cache screenshot checksums from chromedp
var checksums map[string]string

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	var screenshots = c.GetConfig().SupersaaScreenshot
	checksums = make(map[string]string, len(screenshots))

	for {
		processScreenshots(screenshots)

		screenshotInterval := c.GetConfig().ScreenshotInterval
		slog.Info("Will wait for " + screenshotInterval.String())
		time.Sleep(screenshotInterval)
	}
}

func processScreenshots(screenshots []c.SupersaaScreenshot) {
	for _, screenshot := range screenshots {
		screenshotBytes := getScreenshotFromElement(c.GetConfig().WebsiteUrl, screenshot)
		webpageScreenshotSum := computeShasum(screenshotBytes)
		updated := addOrUpdateChecksums(screenshot.Filename, webpageScreenshotSum)

		if updated {
			err := saveImageToGrayScale(screenshot.Filename, screenshotBytes)
			if err != nil {
				slog.Error("Could not save image %s to grayscale. Error: %s", screenshot.Filename, err)
				continue
			}

			err = uploadToPiZero(screenshot.Filename)
			if err != nil {
				slog.Error("Could not upload file %s to %s. Error: %s", screenshot.Filename, c.GetConfig().RemoteURL, err)
			}
			time.Sleep(2 * time.Second)
		}
	}
}

func addOrUpdateChecksums(key string, checksum string) bool {
	slog.Debug("Checksum " + checksum)
	val, ok := checksums[key]
	if ok && val == checksum {
		slog.Debug("Same checksum as before")
		return false
	}

	slog.Debug("Add or update webpage screenshot")
	checksums[key] = checksum
	return true
}

func uploadToPiZero(filename string) error {
	f, err := os.Open(c.GetConfig().ScreenshotDir + filename)
	if err != nil {
		return err
	}
	values := map[string]io.Reader{
		"image":    f,
		"filename": strings.NewReader(filename),
	}

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}

		// Image file
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return err
			}
		} else {
			// Other fields
			if fw, err = w.CreateFormField(key); err != nil {
				return err
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return err
		}

	}

	w.Close()

	req, err := http.NewRequest("POST", c.GetConfig().RemoteURL, &b)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", w.FormDataContentType())

	client := http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("Bad status: %s", res.Status)
	}

	return err
}

func getScreenshotFromElement(websiteUrl string, s c.SupersaaScreenshot) []byte {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoDefaultBrowserCheck,
		chromedp.NoFirstRun,
		chromedp.Headless,
		chromedp.NoSandbox,
		chromedp.Flag("blink-settings", "imagesEnabled=false,cookieEnabled=false,scriptEnabled=false,disableReadingFromCanvas=false"),
	)

	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	if c.GetConfig().UseRemoteAllocator {
		dockerURL := "wss://localhost:9222"
		ctx, cancel = chromedp.NewRemoteAllocator(context.Background(), dockerURL)
		defer cancel()

	} else {
		ctx, cancel = chromedp.NewExecAllocator(context.Background(), opts...)
		defer cancel()
	}

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	var buf []byte

	tasks := chromedp.Tasks{
		chromedp.EmulateViewport(380, 1200),
		chromedp.Navigate(websiteUrl),
		// Block cookie banner
		network.SetBlockedURLS([]string{"https://sak.dnt-userreport.com/sanoma/launcher.js"}),
	}

	for _, a := range s.Actions {
		tasks = append(tasks, chromedp.Evaluate(a, nil))
	}

	tasks = append(tasks, chromedp.Screenshot(s.Selection, &buf, chromedp.NodeVisible))

	err := chromedp.Run(ctx, tasks)
	if err != nil {
		log.Fatal("chromedp.Run", err)
	}

	return buf
}

func computeShasum(buf []byte) string {
	h := sha1.New()
	h.Write(buf)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func saveImageToGrayScale(filename string, buf []byte) error {
	slog.Info("Saving image to grayscale")
	img, _, err := image.Decode(bytes.NewReader(buf))
	if err != nil {
		return err
	}

	var gray = rgbaToGray(img)

	f, err := os.Create(c.GetConfig().ScreenshotDir + filename)
	if err != nil {
		return err
	}
	defer f.Close()

	err = png.Encode(f, gray)
	if err != nil {
		return err
	}

	return nil
}

func rgbaToGray(img image.Image) *image.Gray16 {
	var (
		bounds = img.Bounds()
		gray   = image.NewGray16(bounds)
	)

	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			var rgba = img.At(x, y)
			gray.Set(x, y, rgba)
		}
	}

	return gray
}

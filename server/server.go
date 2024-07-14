package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/quelcom/homeink/config"

	_ "github.com/quelcom/homeink/docs"
	"github.com/quelcom/homeink/kindle"
	slogchi "github.com/samber/slog-chi"
	//httpSwagger "github.com/swaggo/http-swagger/v2"
)

type WaterPayload struct {
	Liters int `json:"liters" example:"42"`
}

// @title Homeink
// @version 1.0
// @description Homeink HTTP server.
func Serve() {
	var port = config.GetConfig().Port

	logger := slog.Default()

	router := chi.NewRouter()
	router.Use(slogchi.New(logger.WithGroup("http")))
	router.Use(middleware.Recoverer)

	router.Route("/api/v1", func(r chi.Router) {
		r.Post("/water", handleWater)
		r.Post("/screenshot", handleScreenshot)
	})

	// router.Get("/swagger/*", httpSwagger.Handler(
	// 	httpSwagger.URL("/swagger/doc.json"),
	// ))

	slog.Info("Starting HTTP server in port " + port)
	log.Fatal(http.ListenAndServe(port, router))
}

// handleWater - Updates water usage
// @Summary Updates water meter in liters
// @Tags Homeink
// @Accept  json
// @Produce  json
// @Param liters body WaterPayload true "Water in liters"
// @Success 200 {object} object{message=string}
// @Failure 400 string http.StatusBadRequest
// @Router /api/v1/water [post]
func handleWater(w http.ResponseWriter, r *http.Request) {

	var water WaterPayload

	err := json.NewDecoder(r.Body).Decode(&water)
	if err != nil {
		http.Error(w, "Problem in parsing request data", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	litersText := fmt.Sprintf("'Water %d L.'", water.Liters)
	err = kindle.Run(kindle.FBInk{Text: litersText, Size: 3, Col: 20, Row: 10}.String())
	if err != nil {
		slog.Error(err.Error())
	}

	response := fmt.Sprintf("Hello World. Liters are %d", water.Liters)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": response})
}

// handleScreenshot - Returns root
// @Summary This API can be used as health check for this application.
// @Tags Homeink
// @Accept  multipart/form-data
// @Produce  json
// @Param image formData []file true "Image to upload"
// @Param filename formData string true "Filename"
// @Success 200 {object} object{message=string}
// @Failure 400 string http.StatusBadRequest
// @Router /api/v1/screenshot [post]
func handleScreenshot(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(0)
	if err != nil {
		http.Error(w, "Problem in parsing request data", http.StatusBadRequest)
		return
	}

	filename := r.FormValue("filename")
	slog.Debug(filename)

	var col, row int
	switch {
	case filename == "supersaa-hourly.png":
		col, row = 2, 22
	case filename == "supersaa-weekly.png":
		col, row = 28, 22
	default:
		http.Error(w, fmt.Sprintf("Filename %q not accepted", filename), http.StatusBadRequest)
		return
	}

	image, err := writeLocalFile(r, "image")
	if err != nil {
		http.Error(w, "Error in writing image to disk", http.StatusInternalServerError)
		return
	}

	kindlePath := "/tmp/" + image
	err = kindle.CopyFile(image, kindlePath)
	if err != nil {
		http.Error(w, "Error in copying image to kindle", http.StatusInternalServerError)
		return
	}
	err = kindle.RenderImage(kindlePath, col, row)
	if err != nil {
		http.Error(w, "Error in rendering image to kindle", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Everything ok"})
}

func writeLocalFile(r *http.Request, key string) (string, error) {
	file, filehandler, err := r.FormFile(key)
	if err != nil {
		return "", errors.New("Error in file")
	}
	defer file.Close()

	localFile, err := os.Create(filehandler.Filename)
	if err != nil {
		return "", errors.New("Could not create file")
	}

	defer localFile.Close()
	if _, err := io.Copy(localFile, file); err != nil {
		return "", errors.New("Could not write to file")
	}

	return localFile.Name(), nil
}

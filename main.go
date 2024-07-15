package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/quelcom/homeink/kindle"
	"github.com/quelcom/homeink/pi"
	"github.com/quelcom/homeink/server"
)

var buildCommit string

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	logger = logger.With("appCommit", buildCommit)
	slog.SetDefault(logger)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		kindle.Disconnect()
		os.Exit(0)
	}()

	kindle.Connect()
	err := kindle.Clear()
	if err != nil {
		slog.Error("Could not clear the screen")
	}
	go kindleClock()
	server.Serve()
}

func kindleClock() {
	for {
		slog.Debug("tick")

		n := time.Now()

		t := fmt.Sprintf("%02d:%02d", n.Hour(), n.Minute())
		err := kindle.Run(kindle.FBInk{Text: t, Size: 11, CenterX: true, Row: 1}.String())
		if err != nil {
			slog.Error(err.Error())
		}

		currentTemp, updated := pi.GetPiTemp()
		if updated {
			err = kindle.Run(kindle.FBInk{Text: "PiZero_temp=" + currentTemp, Size: 3, Col: 2, Row: 10}.String())
			if err != nil {
				slog.Error(err.Error())
			}
		} else {
			slog.Debug("Skip fbink print temp")
		}

		secondsUntilMinuteChange := 60 - n.Second()
		time.Sleep(time.Duration(secondsUntilMinuteChange) * time.Second)
	}
}

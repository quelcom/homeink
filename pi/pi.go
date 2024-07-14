package pi

import (
	"fmt"
	"log/slog"
	"os/exec"
	"runtime"
)

var lastTemp string

func KindleConnected() bool {
	cmd := "lsusb | grep RNDIS"
	err := exec.Command("bash", "-c", cmd).Run()
	return err == nil
}

func GetPiTemp() (string, bool) {
	if runtime.GOOS != "linux" {
		return "35", true
	}

	tempCmd := exec.Command("sudo", "vcgencmd", "measure_temp")
	tempOut, err := tempCmd.Output()
	if err != nil {
		slog.Error("Could not get temperature. Error: " + err.Error())
		return "unknown", true
	}

	currentTemp := string(tempOut)[5:7]
	changed := false

	if currentTemp != lastTemp {
		changed = true
		lastTemp = currentTemp
	}

	slog.Info(fmt.Sprintf("GetPiTemp: current temp %s", currentTemp))
	return currentTemp, changed
}

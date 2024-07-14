package kindle

import (
	"fmt"
	"log"
	"log/slog"
	"strings"

	"github.com/melbahja/goph"
	"github.com/quelcom/homeink/config"
	"github.com/quelcom/homeink/pi"
)

var kindleConnected = pi.KindleConnected()

type FBInk struct {
	Text      string
	Size      int
	Verbose   bool
	CenterX   bool
	CenterY   bool
	Row       int
	Col       int
	Clear     bool
	Invert    bool
	ImagePath string
}

func (fb FBInk) String() string {
	b := strings.Builder{}
	pathToFBInk := config.GetConfig().PathToFBInk
	b.WriteString(pathToFBInk)

	if fb.Clear {
		b.WriteString(" -k")
	}

	if fb.Invert {
		b.WriteString(" --invert")
	}

	if !fb.Verbose {
		b.WriteString(" -q")
	}

	if fb.Size != 0 {
		b.WriteString(fmt.Sprintf(" -S %d", fb.Size))
	}

	if fb.CenterX {
		b.WriteString(" -m")
	} else if fb.Col != 0 {
		b.WriteString(fmt.Sprintf(" -x %d", fb.Col))
	}

	if fb.CenterY {
		b.WriteString(" -M")
	} else if fb.Row != 0 {
		b.WriteString(fmt.Sprintf(" -y %d", fb.Row))
	}

	if fb.Text != "" {
		b.WriteString(fmt.Sprintf(" %s", fb.Text))
	}

	if fb.ImagePath != "" {
		b.WriteString(fmt.Sprintf(" -g file=%s", fb.ImagePath))
	}

	return b.String()
}

var client *goph.Client

func Connect() {
	if kindleConnected {
		kindleIP := config.GetConfig().KindleIP
		kindlePassword := config.GetConfig().KindlePassword

		var err error
		client, err = goph.New("root", kindleIP, goph.Password(kindlePassword))
		if err != nil {
			slog.Error("Cannot get a client")
			log.Fatal(err)
		}
		slog.Debug(fmt.Sprintf("Opened session: %s", client.User()))
	} else {
		slog.Info("Connect: No device attached")
	}
}

func Disconnect() {
	if kindleConnected {
		client.Close()
	} else {
		slog.Info("Disconnect: No device attached")
	}
}

func Run(cmd string) error {
	if kindleConnected {
		_, err := client.Run(cmd)
		return err
	}

	slog.Info(fmt.Sprintf("Run %s: No device attached", cmd))
	return nil

}

func CopyFile(localPath string, remotePath string) error {
	if kindleConnected {
		err := client.Upload(localPath, remotePath)
		return err
	}

	slog.Info(fmt.Sprintf("CopyFile %s %s: No device attached", localPath, remotePath))
	return nil
}

var RenderImage = ActualRenderImage

func ActualRenderImage(path string, col int, row int) error {
	if kindleConnected {
		err := Run(FBInk{ImagePath: path, Col: col, Row: row}.String())
		return err
	}

	slog.Info("RenderImage: No device attached")
	return nil
}

func Clear() error {
	if kindleConnected {
		_ = Run(FBInk{Clear: true, Invert: true}.String())
		err := Run(FBInk{Clear: true}.String())
		return err
	}

	slog.Info("Clear: No device attached")
	return nil
}

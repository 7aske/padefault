package notify

import (
	"fmt"
	"os/exec"
)

const timeout = "1000"

const defaultIcon = "audio-speakers"
const audioVolumeHigh = "audio-volume-high"
const audioVolumeMedium = "audio-volume-medium"
const audioVolumeLow = "audio-volume-low"
const audioVolumeMuted = "audio-off"

func notificationInternal(appName, title, text, iconPath string, hints ...string) {
	args := []string{"-a", appName, "-i", iconPath, title, text, "-t", timeout}
	for _, hint := range hints {
		args = append(args, "-h", hint)
	}

	cmd := exec.Command("notify-send", args...)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error displaying notification:", err)
	}
}

func Default(text string) {
	notificationInternal("padefault", "Padefault", text, defaultIcon)
}

func Volume(text string, volume int) {
	icon := audioVolumeLow
	switch {
	case volume > 66:
		icon = audioVolumeHigh
	case volume > 33:
		icon = audioVolumeMedium
	case volume > 0:
		icon = audioVolumeLow
	case volume == 0:
		icon = audioVolumeMuted
	}

	notificationInternal("padefault", "Volume", fmt.Sprintf(" %s", text), icon, fmt.Sprintf("int:value:%d", volume), "string:synchronous:volume")
}

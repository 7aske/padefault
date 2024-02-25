package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/lawl/pulseaudio"
	"os"
	"padefault/internal/notify"
	"slices"
	"strconv"
	"strings"
)

var cards []pulseaudio.Card
var sinks []pulseaudio.Sink
var defaultSink string
var defaultSource string
var ignoredSinks = []string{
	"alsa_output.usb-Generic_USB_Audio-00.HiFi__SPDIF__sink",
	"alsa_output.usb-Generic_USB_Audio-00.HiFi__Headphones__sink",
}

const maxVolume float64 = 1.5

func sendToggleNotification(newSinkName string) {
	var text string
	for _, sink := range sinks {
		if sink.Name == newSinkName {
			text += fmt.Sprintf("[x] %s\n", sink.Description)
		} else {
			text += fmt.Sprintf("[ ] %s\n", sink.Description)
		}

	}

	notify.Default(text)
}

func PadefaultToggle(client *pulseaudio.Client) error {
	fmt.Println("PadefaultToggle")

	nextSink := -1
	fmt.Println("Default sink", defaultSink)
	for i, sink := range sinks {
		fmt.Println("-- Sink", sink.Name)
		if sink.Name == defaultSink {
			nextSink = i + 1
		}
	}

	if nextSink >= len(sinks) {
		nextSink = 0
	}

	newSink := sinks[nextSink]
	fmt.Println("Setting default sink to", newSink.Name, nextSink)
	err := client.SetDefaultSink(newSink.Name)
	if err != nil {
		return err
	}

	sendToggleNotification(newSink.Name)

	return nil
}

func PadefaultVolume(client *pulseaudio.Client, arg string) error {
	var newVolume float64
	percent := false
	relative := false
	currentVolume, err := client.Volume()
	if err != nil {
		return err
	}

	volumeArg := arg
	if strings.HasSuffix(arg, "%") {
		percent = true
		volumeArg = strings.TrimSuffix(arg, "%")
	}

	if strings.HasPrefix(arg, "+") || strings.HasPrefix(arg, "-") {
		relative = true
	}

	volume, err := strconv.Atoi(volumeArg)
	if err != nil {
		return err
	}

	if percent {
		newVolume = float64(volume) / 100
	}

	if relative {
		newVolume = float64(currentVolume) + float64(volume)/100
	} else {
		newVolume = float64(volume) / 100
	}

	if newVolume > maxVolume {
		newVolume = maxVolume
	} else if newVolume < 0 {
		newVolume = 0
	}

	fmt.Println("Setting volume to", newVolume)
	err = client.SetVolume(float32(newVolume))

	notify.Volume(arg, int(newVolume*100))

	return err
}

func main() {
	parser := parseArgs()

	client, err := pulseaudio.NewClient()
	if err != nil {
		_ = fmt.Errorf("error creating pulseaudio client: %s", err)
		return
	}

	err = setup(client)
	if err != nil {
		_ = fmt.Errorf("error setting up pulseaudio client: %s", err)
		return
	}

	fmt.Println("Cards")
	for _, card := range cards {
		fmt.Println(card.Name)
	}

	fmt.Println("Sinks")
	for _, sink := range sinks {
		fmt.Println(sink.Name)
	}

forLoop:
	for _, command := range parser.GetCommands() {
		if !command.Happened() {
			continue
		}

		switch command.GetName() {
		case "toggle":
			err = PadefaultToggle(client)
			break forLoop
		case "volume":
			arg := command.GetArgs()[1]
			err = PadefaultVolume(client, *arg.GetResult().(*string))
			break forLoop
		}
	}

	if err != nil {
		_ = fmt.Errorf("error running command: %s", err)
	}
}

func setup(client *pulseaudio.Client) (err error) {
	cards, err = client.Cards()
	if err != nil {
		return
	}

	s, err := client.Sinks()
	if err != nil {
		return
	}

	for _, sink := range s {
		if slices.Contains(ignoredSinks, sink.Name) {
			continue
		}

		sinks = append(sinks, sink)
	}

	info, err := client.ServerInfo()
	if err != nil {
		return
	}

	defaultSink = info.DefaultSink
	fmt.Println("Default sink:", defaultSink)
	defaultSource = info.DefaultSource
	fmt.Println("Default source:", defaultSource)

	return
}

func parseArgs() *argparse.Parser {

	parser := argparse.NewParser("padefault",
		"PulseAudio wrapper")

	// Command
	_ = parser.NewCommand("toggle", "Toggle default sink")
	_ = parser.NewCommand("volume", "Set volume").StringPositional(&argparse.Options{
		Required: true,
		Help:     "Volume to set",
		Validate: func(args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("expected 1 argument, got %d", len(args))
			}
			return nil
		},
	})

	err := parser.Parse(os.Args)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, parser.Usage(err))
		os.Exit(1)
	}

	return parser
}

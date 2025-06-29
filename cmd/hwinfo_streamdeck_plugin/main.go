package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	plugin "github.com/shayne/hwinfo-streamdeck/internal/app/hwinfostreamdeckplugin"
)

var port = flag.String("port", "", "The port that should be used to create the WebSocket")
var pluginUUID = flag.String("pluginUUID", "", "A unique identifier string that should be used to register the plugin once the WebSocket is opened")
var registerEvent = flag.String("registerEvent", "", "Registration event")
var info = flag.String("info", "", "A stringified json containing the Stream Deck application information and devices information")

func main() {
	// make sure files are read relative to exe
	err := os.Chdir(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatalf("Unable to chdir: %v", err)
	}

	// Enable debug logging
	appdata := os.Getenv("APPDATA")
	logpath := filepath.Join(appdata, "Elgato/StreamDeck/Plugins/com.exension.hwinfo.sdPlugin/hwinfo.log")
	f, err := os.OpenFile(logpath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("OpenFile Log: %v", err)
	}
	err = f.Truncate(0)
	if err != nil {
		log.Fatalf("Truncate Log: %v", err)
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Fatalf("File Close: %v", err)
		}
	}()
	log.SetOutput(f)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	flag.Parse()

	p, err := plugin.NewPlugin(*port, *pluginUUID, *registerEvent, *info)
	if err != nil {
		log.Fatal("NewPlugin failed:", err)
	}

	err = p.RunForever()
	if err != nil {
		log.Fatal("runForever", err)
	}
}

package hwinfostreamdeckplugin

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-plugin"
	hwsensorsservice "github.com/shayne/hwinfo-streamdeck/pkg/service"
	"github.com/shayne/hwinfo-streamdeck/pkg/graph"
	"github.com/shayne/hwinfo-streamdeck/pkg/streamdeck"
)

// Plugin handles information between HWiNFO and Stream Deck
type Plugin struct {
	c      *plugin.Client
	cmd    *exec.Cmd
	hw     hwsensorsservice.HardwareService
	sd     *streamdeck.StreamDeck
	am     *actionManager
	graphs map[string]*graph.Graph

	appLaunched bool
}

func (p *Plugin) startClient() error {
	cmd := exec.Command("./hwinfo-plugin.exe")
	p.cmd = cmd

	// We're a host. Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  hwsensorsservice.Handshake,
		Plugins:          hwsensorsservice.PluginMap,
		Cmd:             cmd,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		AutoMTLS:        true,
	})

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		return fmt.Errorf("failed to connect to plugin: %w", err)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("hwinfoplugin")
	if err != nil {
		return fmt.Errorf("failed to dispense plugin: %w", err)
	}

	p.c = client
	p.hw = raw.(hwsensorsservice.HardwareService)

	return nil
}

// NewPlugin creates an instance and initializes the plugin
func NewPlugin(port, uuid, event, info string) (*Plugin, error) {
	// Enable logging for debugging
	log.SetOutput(os.Stderr)

	p := &Plugin{
		am:     newActionManager(),
		graphs: make(map[string]*graph.Graph),
	}

	if err := p.startClient(); err != nil {
		return nil, fmt.Errorf("failed to start plugin: %w", err)
	}

	p.sd = streamdeck.NewStreamDeck(port, uuid, event, info)
	return p, nil
}

// RunForever starts the plugin and waits for events, indefinitely
func (p *Plugin) RunForever() error {
	defer func() {
		if p.c != nil {
			p.c.Kill()
		}
	}()

	p.sd.SetDelegate(p)
	p.am.Run(p.updateTiles)

	go func() {
		for {
			if p.c != nil && p.c.Exited() {
				log.Println("Plugin process exited, attempting to restart...")
				if err := p.startClient(); err != nil {
					log.Printf("Failed to restart plugin: %v\n", err)
				}
			}
			time.Sleep(1 * time.Second)
		}
	}()

	err := p.sd.Connect()
	if err != nil {
		return fmt.Errorf("StreamDeck Connect: %v", err)
	}
	defer p.sd.Close()
	p.sd.ListenAndWait()
	return nil
}

func (p *Plugin) getReading(suid string, rid int32) (hwsensorsservice.Reading, error) {
	rbs, err := p.hw.ReadingsForSensorID(suid)
	if err != nil {
		return nil, fmt.Errorf("getReading ReadingsBySensor failed: %v", err)
	}
	for _, r := range rbs {
		if r.ID() == rid {
			return r, nil
		}
	}
	return nil, fmt.Errorf("ReadingID does not exist: %s", suid)
}

// formatWithThousands adds thousands separator to a number string
func formatWithThousands(numStr string) string {
	parts := strings.Split(numStr, ".")
	intPart := parts[0]

	// Handle negative numbers
	sign := ""
	if strings.HasPrefix(intPart, "-") {
		sign = "-"
		intPart = strings.TrimPrefix(intPart, "-")
	}

	// Add commas
	var result strings.Builder
	n := len(intPart)
	for i := 0; i < n; i++ {
		if i > 0 && (n-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteByte(intPart[i])
	}

	// Add back the decimal part if it exists
	if len(parts) > 1 {
		result.WriteRune('.')
		result.WriteString(parts[1])
	}

	return sign + result.String()
}

func (p *Plugin) applyDefaultFormat(v float64, t hwsensorsservice.ReadingType, u string) string {
	// First format the number using standard formatting
	var numStr string
	switch t {
	case hwsensorsservice.ReadingTypeNone:
		numStr = fmt.Sprintf("%.0f", v)
	case hwsensorsservice.ReadingTypeTemp:
		numStr = fmt.Sprintf("%.0f", v)
	case hwsensorsservice.ReadingTypeVolt:
		numStr = fmt.Sprintf("%.0f", v)
	case hwsensorsservice.ReadingTypeFan:
		numStr = fmt.Sprintf("%.0f", v)
	case hwsensorsservice.ReadingTypeCurrent:
		numStr = fmt.Sprintf("%.0f", v)
	case hwsensorsservice.ReadingTypePower:
		numStr = fmt.Sprintf("%.0f", v)
	case hwsensorsservice.ReadingTypeClock:
		numStr = fmt.Sprintf("%.0f", v)
	case hwsensorsservice.ReadingTypeUsage:
		numStr = fmt.Sprintf("%.0f", v)
	case hwsensorsservice.ReadingTypeOther:
		numStr = fmt.Sprintf("%.0f", v)
	default:
		return "Bad Format"
	}

	// Add units based on reading type
	switch t {
	case hwsensorsservice.ReadingTypeTemp:
		if strings.Contains(u, "C") {
			return numStr + " °C"
		} else if strings.Contains(u, "F") {
			return numStr + " °F"
		}
		return numStr + " °C" // Fallback to Celsius
	case hwsensorsservice.ReadingTypeUsage:
		return numStr + u
	default:
		if u != "" {
			return numStr + " " + u
		}
		return numStr
	}
}

// handleFormatString processes custom format strings, supporting %,f for thousands separator
func (p *Plugin) handleFormatString(format string, v float64, t hwsensorsservice.ReadingType, u string) string {
	// Check if the format contains our special thousands separator verb
	if strings.Contains(format, "%,") {
		// Replace %,f or %,d with regular %f and apply thousands separator after
		format = strings.NewReplacer("%,f", "%f", "%,d", "%.0f").Replace(format)
		numStr := fmt.Sprintf(format, v)
		numStr = formatWithThousands(numStr)

		// If the format string doesn't already include the unit, append it
		if !strings.Contains(format, u) {
			switch t {
			case hwsensorsservice.ReadingTypeTemp:
				if strings.Contains(u, "C") {
					return numStr + " °C"
				} else if strings.Contains(u, "F") {
					return numStr + " °F"
				}
				return numStr + " °C" // Fallback to Celsius
			case hwsensorsservice.ReadingTypeUsage:
				return numStr + u
			default:
				if u != "" {
					return numStr + " " + u
				}
			}
		}
		return numStr
	}

	// Regular formatting
	numStr := fmt.Sprintf(format, v)
	// If the format string doesn't already include the unit, append it
	if !strings.Contains(format, u) {
		switch t {
		case hwsensorsservice.ReadingTypeTemp:
			if strings.Contains(u, "C") {
				return numStr + " °C"
			} else if strings.Contains(u, "F") {
				return numStr + " °F"
			}
			return numStr + " °C" // Fallback to Celsius
		case hwsensorsservice.ReadingTypeUsage:
			return numStr + u
		default:
			if u != "" {
				return numStr + " " + u
			}
		}
	}
	return numStr
}

func (p *Plugin) updateTiles(data *actionData) {
	if data.action != "com.exension.hwinfo.reading" {
		log.Printf("Unknown action updateTiles: %s\n", data.action)
		return
	}

	g, ok := p.graphs[data.context]
	if !ok {
		log.Printf("Graph not found for context: %s\n", data.context)
		return
	}

	if !p.appLaunched {
		if !data.settings.InErrorState {
			payload := evStatus{Error: true, Message: "HWiNFO Unavailable"}
			err := p.sd.SendToPropertyInspector("com.exension.hwinfo.reading", data.context, payload)
			if err != nil {
				log.Println("updateTiles SendToPropertyInspector", err)
			}
			data.settings.InErrorState = true
			p.sd.SetSettings(data.context, &data.settings)
		}
		bts, err := ioutil.ReadFile("./launch-hwinfo.png")
		if err != nil {
			log.Printf("Failed to read launch-hwinfo.png: %v\n", err)
			return
		}
		err = p.sd.SetImage(data.context, bts)
		if err != nil {
			log.Printf("Failed to setImage: %v\n", err)
			return
		}
		return
	}

	// show ui on property inspector if in error state
	if data.settings.InErrorState {
		payload := evStatus{Error: false, Message: "show_ui"}
		err := p.sd.SendToPropertyInspector("com.exension.hwinfo.reading", data.context, payload)
		if err != nil {
			log.Println("updateTiles SendToPropertyInspector", err)
		}
		data.settings.InErrorState = false
		p.sd.SetSettings(data.context, &data.settings)
	}

	s := data.settings
	r, err := p.getReading(s.SensorUID, s.ReadingID)
	if err != nil {
		log.Printf("getReading failed: %v\n", err)
		return
	}

	v := r.Value()
	if s.Divisor != "" {
		fdiv := 1.
		fdiv, err := strconv.ParseFloat(s.Divisor, 64)
		if err != nil {
			log.Printf("Failed to parse float: %s\n", s.Divisor)
			return
		}
		v = r.Value() / fdiv
	}
	g.Update(v)
	var text string
	if f := s.Format; f != "" {
		text = p.handleFormatString(f, v, hwsensorsservice.ReadingType(r.TypeI()), r.Unit())
	} else {
		text = p.applyDefaultFormat(v, hwsensorsservice.ReadingType(r.TypeI()), r.Unit())
	}
	g.SetLabelText(1, text)

	b, err := g.EncodePNG()
	if err != nil {
		log.Printf("Failed to encode graph: %v\n", err)
		return
	}

	err = p.sd.SetImage(data.context, b)
	if err != nil {
		log.Printf("Failed to setImage: %v\n", err)
		return
	}
}

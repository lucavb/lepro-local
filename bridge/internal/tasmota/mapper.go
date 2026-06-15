package tasmota

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lucavb/lepro-local/bridge/internal/dp"
)

type Message struct {
	Topic   string
	Payload string
	Retain  bool
}

type State struct {
	Power         bool
	HasPower      bool
	Brightness    int
	HasBrightness bool
	RGB           dp.RGB
	HasRGB        bool
}

func PowerTopic(name string) string       { return "cmnd/" + name + "/POWER" }
func SetPowerTopic(name string) string    { return PowerTopic(name) }
func StatePowerTopic(name string) string  { return "stat/" + name + "/POWER" }
func DimmerTopic(name string) string      { return "cmnd/" + name + "/Dimmer" }
func StateDimmerTopic(name string) string { return "stat/" + name + "/Dimmer" }
func ColorTopic(name string) string       { return "cmnd/" + name + "/Color" }
func StateColorTopic(name string) string  { return "stat/" + name + "/Color" }
func LWTTopic(name string) string         { return "tele/" + name + "/LWT" }

func RetainedMessages(name string, st State) []Message {
	msgs := []Message{
		{Topic: LWTTopic(name), Payload: "Online", Retain: true},
	}
	if st.HasPower {
		payload := "OFF"
		if st.Power {
			payload = "ON"
		}
		msgs = append(msgs, Message{Topic: StatePowerTopic(name), Payload: payload, Retain: true})
	}
	if st.HasBrightness {
		msgs = append(msgs, Message{
			Topic:   StateDimmerTopic(name),
			Payload: strconv.Itoa(dp.Clamp(st.Brightness/10, 0, 100)),
			Retain:  true,
		})
	}
	if st.HasRGB {
		msgs = append(msgs, Message{
			Topic:   StateColorTopic(name),
			Payload: fmt.Sprintf("%d,%d,%d", st.RGB.R, st.RGB.G, st.RGB.B),
			Retain:  true,
		})
	}
	return msgs
}

func ParsePower(payload string) (bool, bool, error) {
	switch strings.ToUpper(strings.TrimSpace(payload)) {
	case "ON", "1", "TRUE":
		return true, true, nil
	case "OFF", "0", "FALSE":
		return false, true, nil
	default:
		return false, false, fmt.Errorf("unsupported power payload %q", payload)
	}
}

func ParseDimmer(payload string) (int, bool, error) {
	n, err := strconv.Atoi(strings.TrimSpace(payload))
	if err != nil {
		return 0, false, err
	}
	if n < 0 || n > 100 {
		return 0, false, fmt.Errorf("dimmer out of range: %d", n)
	}
	return n, true, nil
}

func ParseColor(payload string) (dp.RGB, bool, error) {
	parts := strings.Split(payload, ",")
	if len(parts) != 3 {
		return dp.RGB{}, false, fmt.Errorf("color payload must be R,G,B")
	}
	vals := make([]int, 3)
	for i, part := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return dp.RGB{}, false, err
		}
		vals[i] = dp.Clamp(n, 0, 255)
	}
	return dp.RGB{R: vals[0], G: vals[1], B: vals[2]}, true, nil
}

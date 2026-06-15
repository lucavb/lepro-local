package tasmota

import (
	"testing"

	"github.com/lucavb/lepro-local/bridge/internal/dp"
)

func TestTopics(t *testing.T) {
	if got := PowerTopic("patio-lights"); got != "cmnd/patio-lights/POWER" {
		t.Fatalf("unexpected power topic: %s", got)
	}
	if got := StateColorTopic("patio-lights"); got != "stat/patio-lights/Color" {
		t.Fatalf("unexpected color topic: %s", got)
	}
}

func TestRetainedMessages(t *testing.T) {
	msgs := RetainedMessages("patio-lights", State{
		Power:         true,
		HasPower:      true,
		Brightness:    50,
		HasBrightness: true,
		RGB:           dp.RGB{R: 255, G: 0, B: 0},
		HasRGB:        true,
	})
	if len(msgs) != 4 {
		t.Fatalf("expected 4 retained messages, got %d", len(msgs))
	}
}

func TestParseColor(t *testing.T) {
	rgb, ok, err := ParseColor("255,0,128")
	if err != nil || !ok {
		t.Fatalf("parse color: %v", err)
	}
	if rgb.R != 255 || rgb.G != 0 || rgb.B != 128 {
		t.Fatalf("unexpected rgb: %+v", rgb)
	}
}

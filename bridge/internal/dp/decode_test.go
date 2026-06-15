package dp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func loadInnerFixture(t *testing.T, name string) Payload {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "tests", "fixtures", "dp", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	var fixture map[string]any
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("unmarshal fixture %s: %v", name, err)
	}
	inner, ok := fixture["lepro_set"].(map[string]any)
	if !ok {
		t.Fatalf("fixture %s missing lepro_set", name)
	}
	return Payload(inner["d"].(map[string]any))
}

func TestEnvelopeRoundTrip(t *testing.T) {
	outer := Wrap(Power(true))
	raw := JSONText(outer)
	parsed, err := EnvelopeFromJSON([]byte(raw))
	if err != nil {
		t.Fatalf("parse envelope: %v", err)
	}
	if JSONText(parsed) != JSONText(Power(true)) {
		t.Fatalf("round trip mismatch\n got: %s\nwant: %s", JSONText(parsed), JSONText(Power(true)))
	}
}

func TestDecodeColorFixtures(t *testing.T) {
	for _, name := range []string{"color_red.json", "color_green.json", "color_blue.json"} {
		inner := loadInnerFixture(t, name)
		state, err := DecodeState(inner)
		if err != nil {
			t.Fatalf("decode %s: %v", name, err)
		}
		if !state.HasRGB {
			t.Fatalf("decode %s: expected RGB", name)
		}
		reencoded := ColorRGB(state.RGB.R, state.RGB.G, state.RGB.B, state.Brightness)
		if JSONText(reencoded) != JSONText(inner) {
			t.Fatalf("round trip mismatch for %s\n got: %s\nwant: %s", name, JSONText(reencoded), JSONText(inner))
		}
	}
}

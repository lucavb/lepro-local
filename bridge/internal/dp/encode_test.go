package dp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func loadFixture(t *testing.T, name string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "tests", "fixtures", "dp", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	var fixture map[string]any
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("unmarshal fixture %s: %v", name, err)
	}
	return fixture
}

func requireInner(t *testing.T, got Payload, fixture map[string]any) {
	t.Helper()
	want := fixture["lepro_inner"]
	if got == nil {
		t.Fatalf("got nil payload")
	}
	if JSONText(got) != JSONText(want) {
		t.Fatalf("payload mismatch\n got: %s\nwant: %s", JSONText(got), JSONText(want))
	}
}

func TestPowerFixtures(t *testing.T) {
	requireInner(t, Power(true), loadFixture(t, "power_on.json"))
	requireInner(t, Power(false), loadFixture(t, "power_off.json"))
}

func TestColorFixtures(t *testing.T) {
	requireInner(t, ColorRGB(255, 0, 0, 1000), loadFixture(t, "color_red.json"))
	requireInner(t, ColorRGB(0, 255, 0, 1000), loadFixture(t, "color_green.json"))
	requireInner(t, ColorRGB(0, 0, 255, 1000), loadFixture(t, "color_blue.json"))
}

func TestDimmerFixtureWithState(t *testing.T) {
	got := Dimmer(50, &State{
		HasRGB: true,
		RGB:    RGB{R: 0, G: 255, B: 0},
	})
	requireInner(t, got, loadFixture(t, "brightness_50.json"))
}

func TestWrap(t *testing.T) {
	got := Wrap(Power(true))
	want := loadFixture(t, "power_on.json")["lepro_set"]
	if JSONText(got) != JSONText(want) {
		t.Fatalf("wrap mismatch\n got: %s\nwant: %s", JSONText(got), JSONText(want))
	}
}

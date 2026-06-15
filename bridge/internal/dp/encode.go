package dp

import (
	"encoding/json"
	"fmt"
	"math"
)

type RGB struct {
	R int
	G int
	B int
}

type State struct {
	Power         bool
	HasPower      bool
	Brightness    int
	HasBrightness bool
	RGB           RGB
	HasRGB        bool
}

type Payload map[string]any

func Clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func JSONText(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func Wrap(inner Payload) Payload {
	return Payload{"d": inner}
}

func Power(on bool) Payload {
	if on {
		return Payload{"d1": 1}
	}
	return Payload{"d1": 0}
}

func HSVHex(hue, sat, val int) string {
	return fmt.Sprintf("%04X%04X%04X", Clamp(hue, 0, 359), Clamp(sat, 0, 1000), Clamp(val, 0, 1000))
}

func RGBToHSV(r, g, b int) (hue, sat, val int) {
	rf := float64(Clamp(r, 0, 255)) / 255.0
	gf := float64(Clamp(g, 0, 255)) / 255.0
	bf := float64(Clamp(b, 0, 255)) / 255.0
	maxc := math.Max(rf, math.Max(gf, bf))
	minc := math.Min(rf, math.Min(gf, bf))
	delta := maxc - minc

	val = int(math.Round(maxc * 1000))
	if delta == 0 {
		return 0, 0, val
	}

	sat = int(math.Round((delta / maxc) * 1000))

	var h float64
	switch maxc {
	case rf:
		h = math.Mod((gf-bf)/delta, 6)
	case gf:
		h = ((bf-rf)/delta + 2)
	default:
		h = ((rf-gf)/delta + 4)
	}
	hue = int(math.Round(h * 60))
	if hue < 0 {
		hue += 360
	}
	hue %= 360
	return hue, sat, val
}

func HSVToRGB(hue, sat, val int) (r, g, b int) {
	h := float64(((hue%360)+360)%360) / 60.0
	s := float64(Clamp(sat, 0, 1000)) / 1000.0
	v := float64(Clamp(val, 0, 1000)) / 1000.0
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h, 2)-1))
	m := v - c

	var rf, gf, bf float64
	switch {
	case h < 1:
		rf, gf, bf = c, x, 0
	case h < 2:
		rf, gf, bf = x, c, 0
	case h < 3:
		rf, gf, bf = 0, c, x
	case h < 4:
		rf, gf, bf = 0, x, c
	case h < 5:
		rf, gf, bf = x, 0, c
	default:
		rf, gf, bf = c, 0, x
	}

	r = int(math.Round((rf + m) * 255))
	g = int(math.Round((gf + m) * 255))
	b = int(math.Round((bf + m) * 255))
	return Clamp(r, 0, 255), Clamp(g, 0, 255), Clamp(b, 0, 255)
}

func ColorRGB(r, g, b, brightness int) Payload {
	hue, sat, val := RGBToHSV(r, g, b)
	return Payload{
		"d1": 1,
		"d2": 1,
		"d3": Clamp(brightness, 10, 1000),
		"d5": HSVHex(hue, sat, val),
	}
}

func Dimmer(brightness int, current *State) Payload {
	payload := Payload{
		"d1": 1,
		"d3": Clamp(brightness*10, 10, 1000),
	}
	if current != nil && current.HasRGB {
		hue, sat, _ := RGBToHSV(current.RGB.R, current.RGB.G, current.RGB.B)
		payload["d2"] = 1
		payload["d5"] = HSVHex(hue, sat, Clamp(brightness*10, 10, 1000))
	}
	return payload
}

func WhiteMode(brightness, temp int) Payload {
	return Payload{
		"d1": 1,
		"d2": 0,
		"d3": Clamp(brightness, 10, 1000),
		"d4": Clamp(temp, 0, 1000),
	}
}

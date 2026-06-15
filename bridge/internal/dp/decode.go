package dp

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func EnvelopeFromJSON(raw []byte) (Payload, error) {
	var outer map[string]any
	if err := json.Unmarshal(raw, &outer); err != nil {
		return nil, err
	}
	return OuterToInner(outer)
}

func OuterToInner(outer map[string]any) (Payload, error) {
	if outer == nil {
		return nil, fmt.Errorf("empty payload")
	}
	v, ok := outer["d"]
	if !ok {
		return nil, fmt.Errorf("missing d envelope")
	}
	inner, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("d envelope must be object")
	}
	return Payload(inner), nil
}

func ParseInnerJSON(raw []byte) (Payload, error) {
	var inner map[string]any
	if err := json.Unmarshal(raw, &inner); err != nil {
		return nil, err
	}
	return Payload(inner), nil
}

func DecodeState(inner Payload) (State, error) {
	var st State
	if inner == nil {
		return st, fmt.Errorf("nil inner payload")
	}

	if v, ok := inner["d1"]; ok {
		on, err := asBoolish(v)
		if err != nil {
			return st, fmt.Errorf("d1: %w", err)
		}
		st.Power = on
		st.HasPower = true
	}
	if v, ok := inner["d3"]; ok {
		n, err := asInt(v)
		if err != nil {
			return st, fmt.Errorf("d3: %w", err)
		}
		st.Brightness = Clamp(n, 0, 1000)
		st.HasBrightness = true
	}
	if v, ok := inner["d5"]; ok {
		s, ok := v.(string)
		if !ok {
			return st, fmt.Errorf("d5: expected string")
		}
		r, g, b, err := DecodeHSVHex(s)
		if err != nil {
			return st, fmt.Errorf("d5: %w", err)
		}
		st.RGB = RGB{R: r, G: g, B: b}
		st.HasRGB = true
	}
	return st, nil
}

func DecodeHSVHex(value string) (r, g, b int, err error) {
	if len(value) != 12 {
		return 0, 0, 0, fmt.Errorf("expected 12 hex chars, got %d", len(value))
	}
	hue, err := strconv.ParseInt(value[:4], 16, 32)
	if err != nil {
		return 0, 0, 0, err
	}
	sat, err := strconv.ParseInt(value[4:8], 16, 32)
	if err != nil {
		return 0, 0, 0, err
	}
	val, err := strconv.ParseInt(value[8:12], 16, 32)
	if err != nil {
		return 0, 0, 0, err
	}
	rr, gg, bb := HSVToRGB(int(hue), int(sat), int(val))
	return rr, gg, bb, nil
}

func asInt(v any) (int, error) {
	switch x := v.(type) {
	case float64:
		return int(x), nil
	case float32:
		return int(x), nil
	case int:
		return x, nil
	case int64:
		return int(x), nil
	case json.Number:
		n, err := x.Int64()
		return int(n), err
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(x))
		return n, err
	default:
		return 0, fmt.Errorf("unsupported type %T", v)
	}
}

func asBoolish(v any) (bool, error) {
	switch x := v.(type) {
	case bool:
		return x, nil
	case float64:
		return x != 0, nil
	case int:
		return x != 0, nil
	case int64:
		return x != 0, nil
	case string:
		switch strings.ToLower(strings.TrimSpace(x)) {
		case "1", "true", "on", "yes":
			return true, nil
		case "0", "false", "off", "no":
			return false, nil
		default:
			return false, fmt.Errorf("unsupported boolean string %q", x)
		}
	default:
		return false, fmt.Errorf("unsupported type %T", v)
	}
}

func HSVHexFromRGB(r, g, b int) string {
	h, s, v := RGBToHSV(r, g, b)
	return HSVHex(h, s, v)
}

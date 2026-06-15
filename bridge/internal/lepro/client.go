package lepro

import (
	"encoding/json"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/lucavb/lepro-local/bridge/internal/dp"
)

type Client struct {
	MQTT     mqtt.Client
	DeviceID string
}

func SetTopic(deviceID string) string    { return "le/" + deviceID + "/prp/set" }
func ReportTopic(deviceID string) string { return "le/" + deviceID + "/prp/rpt" }

func (c Client) PublishSet(payload dp.Payload) error {
	raw, err := json.Marshal(dp.Wrap(payload))
	if err != nil {
		return err
	}
	token := c.MQTT.Publish(SetTopic(c.DeviceID), 0, false, raw)
	token.Wait()
	return token.Error()
}

func (c Client) SubscribeReports(handler mqtt.MessageHandler) error {
	token := c.MQTT.Subscribe(ReportTopic(c.DeviceID), 0, handler)
	token.Wait()
	return token.Error()
}

func DecodeReport(msg mqtt.Message) (dp.State, error) {
	inner, err := dp.EnvelopeFromJSON(msg.Payload())
	if err != nil {
		return dp.State{}, fmt.Errorf("decode envelope: %w", err)
	}
	return dp.DecodeState(inner)
}

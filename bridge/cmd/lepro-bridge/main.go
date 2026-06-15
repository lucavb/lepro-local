package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/lucavb/lepro-local/bridge/internal/config"
	"github.com/lucavb/lepro-local/bridge/internal/dp"
	"github.com/lucavb/lepro-local/bridge/internal/lepro"
	"github.com/lucavb/lepro-local/bridge/internal/tasmota"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

type bridgeState struct {
	mu sync.Mutex

	power         bool
	hasPower      bool
	brightness    int
	hasBrightness bool
	rgb           dp.RGB
	hasRGB        bool
}

func (s *bridgeState) snapshot() tasmota.State {
	s.mu.Lock()
	defer s.mu.Unlock()
	return tasmota.State{
		Power:         s.power,
		HasPower:      s.hasPower,
		Brightness:    s.brightness,
		HasBrightness: s.hasBrightness,
		RGB:           s.rgb,
		HasRGB:        s.hasRGB,
	}
}

func (s *bridgeState) applyPower(on bool) tasmota.State {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.power = on
	s.hasPower = true
	return tasmota.State{
		Power:         s.power,
		HasPower:      s.hasPower,
		Brightness:    s.brightness,
		HasBrightness: s.hasBrightness,
		RGB:           s.rgb,
		HasRGB:        s.hasRGB,
	}
}

func (s *bridgeState) applyDimmer(level int) tasmota.State {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.brightness = level
	s.hasBrightness = true
	return tasmota.State{
		Power:         s.power,
		HasPower:      s.hasPower,
		Brightness:    s.brightness,
		HasBrightness: s.hasBrightness,
		RGB:           s.rgb,
		HasRGB:        s.hasRGB,
	}
}

func (s *bridgeState) applyRGB(rgb dp.RGB) tasmota.State {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rgb = rgb
	s.hasRGB = true
	s.power = true
	s.hasPower = true
	if !s.hasBrightness {
		s.brightness = 100
		s.hasBrightness = true
	}
	return tasmota.State{
		Power:         s.power,
		HasPower:      s.hasPower,
		Brightness:    s.brightness,
		HasBrightness: s.hasBrightness,
		RGB:           s.rgb,
		HasRGB:        s.hasRGB,
	}
}

func (s *bridgeState) applyReport(st dp.State) tasmota.State {
	s.mu.Lock()
	defer s.mu.Unlock()
	if st.HasPower {
		s.power = st.Power
		s.hasPower = true
	}
	if st.HasBrightness {
		s.brightness = dp.Clamp(st.Brightness/10, 0, 100)
		s.hasBrightness = true
	}
	if st.HasRGB {
		s.rgb = st.RGB
		s.hasRGB = true
	}
	return tasmota.State{
		Power:         s.power,
		HasPower:      s.hasPower,
		Brightness:    s.brightness,
		HasBrightness: s.hasBrightness,
		RGB:           s.rgb,
		HasRGB:        s.hasRGB,
	}
}

func run() error {
	cfgPath := flag.String("config", os.Getenv("LEPRO_LAB_CONFIG"), "path to lepro-lab.toml (optional)")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		return err
	}

	state := &bridgeState{}

	var home mqtt.Client
	leproTLS, err := cfg.Bridge.TLSConfig()
	if err != nil {
		return err
	}

	leproClient, err := connectMQTT(cfg.Bridge.LeproBroker, cfg.Device.FriendlyName+"-lepro", leproTLS, func(c mqtt.Client) {
		token := c.Subscribe(lepro.ReportTopic(cfg.Device.ID), 0, func(_ mqtt.Client, msg mqtt.Message) {
			report, err := lepro.DecodeReport(msg)
			if err != nil {
				log.Printf("decode report: %v", err)
				return
			}
			publishHomeState(home, cfg.Device.FriendlyName, state.applyReport(report))
		})
		token.Wait()
		if err := token.Error(); err != nil {
			log.Printf("subscribe lepro reports: %v", err)
		}
	})
	if err != nil {
		return fmt.Errorf("connect lepro broker: %w", err)
	}

	homeClient, err := connectMQTT(cfg.Bridge.HomeBroker, cfg.Device.FriendlyName+"-home", nil, func(c mqtt.Client) {
		subscribeHomeTopics(c, cfg.Device.FriendlyName, state, lepro.Client{MQTT: leproClient, DeviceID: cfg.Device.ID})
		publishHomeState(c, cfg.Device.FriendlyName, state.snapshot())
	})
	if err != nil {
		return fmt.Errorf("connect home broker: %w", err)
	}
	home = homeClient

	if token := home.Connect(); !token.WaitTimeout(15*time.Second) || token.Error() != nil {
		return fmt.Errorf("home broker connect: %w", token.Error())
	}
	if token := leproClient.Connect(); !token.WaitTimeout(15*time.Second) || token.Error() != nil {
		return fmt.Errorf("lepro broker connect: %w", token.Error())
	}

	log.Printf("lepro-bridge running for %s (%s)", cfg.Device.FriendlyName, cfg.Device.ID)

	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	<-term

	publishHomeOffline(home, cfg.Device.FriendlyName)
	if home.IsConnected() {
		home.Disconnect(250)
	}
	if leproClient.IsConnected() {
		leproClient.Disconnect(250)
	}
	return nil
}

func connectMQTT(rawURL, clientID string, tlsCfg *tls.Config, onConnect mqtt.OnConnectHandler) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions().AddBroker(rawURL)
	opts.SetClientID(clientID)
	opts.SetCleanSession(false)
	opts.SetAutoReconnect(true)
	opts.SetOrderMatters(false)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(5 * time.Second)
	if tlsCfg != nil {
		opts.SetTLSConfig(tlsCfg)
	}
	if onConnect != nil {
		opts.SetOnConnectHandler(onConnect)
	}

	client := mqtt.NewClient(opts)
	return client, nil
}

func subscribeHomeTopics(c mqtt.Client, name string, state *bridgeState, leproClient lepro.Client) {
	subscribe := func(topic string, handler mqtt.MessageHandler) {
		token := c.Subscribe(topic, 0, handler)
		token.Wait()
		if err := token.Error(); err != nil {
			log.Printf("subscribe %s: %v", topic, err)
		}
	}

	prefix := "cmnd/" + name + "/"
	subscribe(prefix+"POWER", func(_ mqtt.Client, msg mqtt.Message) {
		on, ok, err := tasmota.ParsePower(string(msg.Payload()))
		if err != nil || !ok {
			log.Printf("power command: %v", err)
			return
		}
		snapshot := state.applyPower(on)
		publishHomeState(c, name, snapshot)
		if err := leproClient.PublishSet(dp.Power(on)); err != nil {
			log.Printf("publish lepro power: %v", err)
		}
	})

	subscribe(prefix+"Dimmer", func(_ mqtt.Client, msg mqtt.Message) {
		level, ok, err := tasmota.ParseDimmer(string(msg.Payload()))
		if err != nil || !ok {
			log.Printf("dimmer command: %v", err)
			return
		}
		snapshot := state.applyDimmer(level)
		publishHomeState(c, name, snapshot)
		payload := dp.Dimmer(level, &dp.State{HasRGB: snapshot.HasRGB, RGB: snapshot.RGB})
		if err := leproClient.PublishSet(payload); err != nil {
			log.Printf("publish lepro dimmer: %v", err)
		}
	})

	subscribe(prefix+"Color", func(_ mqtt.Client, msg mqtt.Message) {
		rgb, ok, err := tasmota.ParseColor(string(msg.Payload()))
		if err != nil || !ok {
			log.Printf("color command: %v", err)
			return
		}
		snapshot := state.applyRGB(rgb)
		publishHomeState(c, name, snapshot)
		if err := leproClient.PublishSet(dp.ColorRGB(rgb.R, rgb.G, rgb.B, snapshot.Brightness*10)); err != nil {
			log.Printf("publish lepro color: %v", err)
		}
	})
}

func publishHomeState(c mqtt.Client, name string, st tasmota.State) {
	if c == nil || !c.IsConnected() {
		return
	}
	for _, msg := range tasmota.RetainedMessages(name, st) {
		token := c.Publish(msg.Topic, 0, msg.Retain, msg.Payload)
		token.Wait()
		if err := token.Error(); err != nil {
			log.Printf("publish %s: %v", msg.Topic, err)
		}
	}
}

func publishHomeOffline(c mqtt.Client, name string) {
	if c == nil || !c.IsConnected() {
		return
	}
	token := c.Publish(tasmota.LWTTopic(name), 0, true, "Offline")
	token.Wait()
}

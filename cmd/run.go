// SPDX-FileCopyrightText: 2025 Jacques Supcik <jacques.supcik@hefr.ch>
//
// SPDX-License-Identifier: MIT

package cmd

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"log"

	"math"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"log/slog"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	reboot   = "reboot"
	shutdown = "shutdown"
)

func getUniqueId() (addr string) {
	interfaces, err := net.Interfaces()
	var mac net.HardwareAddr = nil
	if err == nil {
		for _, i := range interfaces {
			if i.Flags&net.FlagUp != 0 && !bytes.Equal(i.HardwareAddr, nil) {
				mac = i.HardwareAddr
				break
			}
		}
	}
	if len(mac) == 0 {
		mac = make([]byte, 6)
		_, err = rand.Read(mac)
		if err != nil {
			log.Print("Error reading random bytes: ", err)
			return "raspi-off-unknown"
		}
	}

	const hexDigit = "0123456789abcdef"
	buf := make([]byte, 0, len(mac)*3-1)
	for _, b := range mac {
		buf = append(buf, hexDigit[b>>4])
		buf = append(buf, hexDigit[b&0xF])
	}
	return "raspi-off-" + string(buf)

}

func onMessageReceived(client mqtt.Client, msg mqtt.Message) {
	arg := string(msg.Payload())
	if arg == "" {
		arg = "now"
	}
	var cmd *exec.Cmd
	if strings.HasSuffix(msg.Topic(), reboot) {
		slog.Info("Rebooting")
		cmd = exec.Command("shutdown", "-r", arg)
		slog.Info(fmt.Sprint("Reboot command: ", cmd))
	} else if strings.HasSuffix(msg.Topic(), shutdown) {
		slog.Info("Shutting down")
		cmd = exec.Command("shutdown", "-h", arg)
		slog.Info(fmt.Sprint("Shutdown command: ", cmd))
	}
	if err := cmd.Run(); err != nil {
		log.Print("Error: ", err)
	}
}

func run() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	if Debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else if Verbose {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	} else {
		slog.SetLogLoggerLevel(slog.LevelWarn)
	}

	slog.Info("Starting raspi-off service")
	c := make(chan os.Signal, 1)

	mqtt.ERROR = NewMqttLogger(slog.LevelError, slog.Default())
	mqtt.CRITICAL = NewMqttLogger(slog.LevelError, slog.Default())
	mqtt.WARN = NewMqttLogger(slog.LevelWarn, slog.Default())
	mqtt.DEBUG = NewMqttLogger(slog.LevelDebug, slog.Default())

	id := getUniqueId()

	slog.Info(fmt.Sprint("MQTT Client ID: ", id))
	slog.Info(fmt.Sprint("MQTT Server: ", Server))

	opts := mqtt.NewClientOptions().AddBroker(Server).SetClientID(id).SetCleanSession(true)
	opts.SetKeepAlive(2 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	opts.ConnectTimeout = 10 * time.Second
	opts.SetConnectRetry(true)
	opts.SetMaxReconnectInterval(math.MaxInt64)

	opts.OnConnect = func(c mqtt.Client) {
		slog.Info("Connected")
		t := BaseTopic + "/" + reboot
		slog.Info(fmt.Sprint("Subscribing to: ", t))
		if token := c.Subscribe(t, byte(QoS), onMessageReceived); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
		t = BaseTopic + "/" + shutdown
		slog.Info(fmt.Sprint("Subscribing to: ", t))
		if token := c.Subscribe(BaseTopic+"/"+shutdown, byte(QoS), onMessageReceived); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	<-c
}

// SPDX-FileCopyrightText: 2025 Jacques Supcik <jacques.supcik@hefr.ch>
//
// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Verbose bool
var Debug bool
var Server string
var BaseTopic string
var QoS int

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "raspi-off",
	Short: "Remote Shutdown and Reboot service for Raspberry Pi",
	Long: `raspi-off is a service that allows you to remotely shutdown or reboot your Raspberry Pi.
It listens for MQTT messages and executes the appropriate command based on the topic.`,

	Run: func(cmd *cobra.Command, args []string) {
		run()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.Flags().BoolVarP(&Debug, "debug", "d", false, "debug output")
	rootCmd.Flags().StringVarP(&Server, "server", "s", "tcp://mqtt.local:1883", "The full URL of the MQTT server to connect to")
	rootCmd.Flags().StringVarP(&BaseTopic, "base-topic", "t", "raspi-off", "Base topic to subscribe to")
	rootCmd.Flags().IntVarP(&QoS, "qos", "q", 1, "The QoS to subscribe to messages at")
	rootCmd.Version = version
}

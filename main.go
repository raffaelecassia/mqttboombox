package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
)

var bb boombox

func quitme() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	bb.close()
	os.Exit(0)
}

func main() {
	// cmdline parsing
	brokerURL := flag.String("broker", "tcp://localhost:1883", "MQTT broker URL")
	isBinary := flag.Bool("binary", false, "binary data mode")
	isCounter := flag.Bool("counter", false, "print messages counter")
	isFastforward := flag.Bool("ff", false, "fastforward, skip messages timing")
	clientID := flag.String("clientid", "", "client id")
	username := flag.String("username", "", "username")
	password := flag.String("password", "", "password")
	flag.Parse()
	topics := flag.Args()

	bb = New(*brokerURL, topics, *isBinary, *isCounter, *isFastforward)

	if *clientID != "" {
		bb.options.SetClientID(*clientID)
	}
	if *username != "" {
		bb.options.SetUsername(*username)
	}
	if *password != "" {
		bb.options.SetPassword(*password)
	}

	// connecting
	bb.mqttConnect()

	go quitme()

	// checks for stdin
	stat, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}
	stdinMode := stat.Mode()

	if (stdinMode&os.ModeNamedPipe) != 0 || stdinMode.IsRegular() {
		bb.play()
	} else {
		bb.record()
	}
}

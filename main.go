package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
)

var bb boombox

// commandline flag parsing
func flagParse() (brokerURL string, topics []string, isBinary bool, isCounter bool, isFastforward bool) {
	flag.StringVar(&brokerURL, "broker", "tcp://localhost:1883", "MQTT broker URL")
	flag.BoolVar(&isBinary, "binary", false, "binary data mode")
	flag.BoolVar(&isCounter, "counter", false, "print messages counter")
	flag.BoolVar(&isFastforward, "ff", false, "fastforward, skip messages timing")
	flag.Parse()
	topics = flag.Args()
	return
}

func quitme() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	bb.close()
	os.Exit(0)
}

func main() {
	// cmdline parsing
	brokerURL, topics, isBinary, isCounter, isFastforward := flagParse()

	bb = boombox{
		brokerURL:     brokerURL,
		topics:        topics,
		isBinary:      isBinary,
		isCounter:     isCounter,
		isFastforward: isFastforward}

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

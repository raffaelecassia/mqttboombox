package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	recordDataSep string = "|"
)

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

// MQTT message wrapper
type message struct {
	time time.Time
	msg  MQTT.Message
}

// record buffered channel
var recordchan = make(chan message, 100)

// default mqtt message handler
var onMessage MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	recordchan <- message{time.Now(), msg}
}

// connects to the mqtt broker
func mqttConnect(brokerURL string) MQTT.Client {
	opts := MQTT.NewClientOptions().AddBroker(brokerURL)
	//~ opts.SetClientID("mqttboombox" + strconv.FormatInt(time.Now().Unix(), 16))
	opts.SetDefaultPublishHandler(onMessage)
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	return client
}

// subscribe to a list of topics
func mqttSubscribe(client MQTT.Client, topics []string) {
	// TODO replace with SubscribeMultiple
	for _, t := range topics {
		if token := client.Subscribe(t, 0, nil); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
}

// RECORD!
func record(client MQTT.Client, topics []string, isBinary bool, isCounter bool) {
	// checks topics list
	if len(topics) == 0 {
		//    exit?
		// fmt.Fprintln(os.Stderr, "You must specify at least one topic!")
		// os.Exit(1)
		//    or default all topics?
		topics = []string{"#"}
	}

	lasttime := time.Now()

	mqttSubscribe(client, topics)

	fmt.Fprintln(os.Stderr, "INPUT TOPICS", topics)
	var (
		count   int
		m       message
		elapsed time.Duration
		payload string
	)
	for {
		m = <-recordchan

		elapsed = m.time.Sub(lasttime)
		lasttime = m.time

		if isBinary {
			payload = base64.StdEncoding.EncodeToString(m.msg.Payload())
		} else {
			payload = string(m.msg.Payload())
		}
		fmt.Printf("%s%s%s%s%s\n",
			elapsed,
			recordDataSep,
			m.msg.Topic(),
			recordDataSep,
			payload)

		if isCounter {
			count++
			fmt.Fprintf(os.Stderr, "\rREC  #%d", count)
		}
	}
}

// PLAY!
func play(client MQTT.Client, isBinary bool, isCounter bool, isFastforward bool) {
	var (
		count   int
		data    []string
		payload []byte
		token   MQTT.Token
		elapsed time.Duration
		err     error
		err2    error
	)
	// reads from stdin
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if len(scanner.Text()) == 0 {
			continue
		}
		// split the line
		data = strings.Split(scanner.Text(), recordDataSep)
		// parse elapsed
		elapsed, err = time.ParseDuration(data[0])
		if err != nil {
			panic(err)
		}

		if isBinary {
			// parse payload
			payload, err2 = base64.StdEncoding.DecodeString(data[2])
			if err2 != nil {
				panic(err2)
			}
		} else {
			payload = []byte(data[2])
		}

		if !isFastforward {
			time.Sleep(elapsed)
		}

		// sends
		if token = client.Publish(data[1], 0, false, payload); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}

		count++
		if isCounter {
			fmt.Fprintf(os.Stderr, "\rPLAY #%d", count)
		}
	}
	if err = scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stderr, "\rEND  #%d\n", count)
}

func quitme(client MQTT.Client) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	client.Disconnect(0)
	fmt.Fprintln(os.Stderr, "\rDISCONNECTED")
	os.Exit(0)
}

func main() {
	// cmdline parsing
	brokerURL, topics, isBinary, isCounter, isFastforward := flagParse()
	// connecting
	client := mqttConnect(brokerURL)

	go quitme(client)

	// checks for stdin
	stat, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}
	stdinMode := stat.Mode()

	if (stdinMode&os.ModeNamedPipe) != 0 || stdinMode.IsRegular() {
		fmt.Fprintln(os.Stderr, "PLAYBACK to", brokerURL)
		play(client, isBinary, isCounter, isFastforward)
	} else {
		fmt.Fprintln(os.Stderr, "RECORDING from", brokerURL)
		record(client, topics, isBinary, isCounter)
	}
}

package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	recordDataSep string = "|"
)

type boombox struct {
	brokerURL     string
	topics        []string
	isBinary      bool
	isCounter     bool
	isFastforward bool
	mqttclient    MQTT.Client
	options       *MQTT.ClientOptions
	recordchan    chan message
}

// MQTT message wrapper
type message struct {
	time time.Time
	msg  MQTT.Message
}

func New(brokerURL string, topics []string, isBinary bool, isCounter bool, isFastforward bool) boombox {
	o := boombox{
		brokerURL:     brokerURL,
		topics:        topics,
		isBinary:      isBinary,
		isCounter:     isCounter,
		isFastforward: isFastforward,
		options:       MQTT.NewClientOptions(),
	}
	return o
}

// connects to the mqtt broker
func (boombox *boombox) mqttConnect() {
	// options
	boombox.options.AddBroker(boombox.brokerURL)
	// opts.SetClientID("mqttboombox" + strconv.FormatInt(time.Now().Unix(), 16))
	boombox.options.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		boombox.recordchan <- message{time.Now(), msg}
	})
	boombox.mqttclient = MQTT.NewClient(boombox.options)
	if token := boombox.mqttclient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
}

// subscribe to the list of topics
func (boombox *boombox) mqttSubscribe() {
	t := make(map[string]byte)
	for _, topic := range boombox.topics {
		t[topic] = 0
	}
	if token := boombox.mqttclient.SubscribeMultiple(t, nil); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
}

// RECORD!
func (boombox *boombox) record() {
	fmt.Fprintln(os.Stderr, "RECORDING from", boombox.brokerURL)
	var (
		count    int
		m        message
		payload  string
		elapsed  time.Duration
		lasttime = time.Now()
	)

	// record buffered channel
	boombox.recordchan = make(chan message, 100)

	// checks topics list and subscribe
	if len(boombox.topics) == 0 {
		// default to all topics
		boombox.topics = []string{"#"}
	}
	boombox.mqttSubscribe()

	fmt.Fprintln(os.Stderr, "INPUT TOPICS", boombox.topics)

	for {
		m = <-boombox.recordchan

		elapsed = m.time.Sub(lasttime)
		lasttime = m.time

		if boombox.isBinary {
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

		if boombox.isCounter {
			count++
			fmt.Fprintf(os.Stderr, "\rREC  #%d", count)
		}
	}
}

// PLAY!
func (boombox *boombox) play() {
	fmt.Fprintln(os.Stderr, "PLAYBACK to", boombox.brokerURL)
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

		if boombox.isBinary {
			// parse payload
			payload, err2 = base64.StdEncoding.DecodeString(data[2])
			if err2 != nil {
				panic(err2)
			}
		} else {
			payload = []byte(data[2])
		}

		if !boombox.isFastforward {
			time.Sleep(elapsed)
		}

		// sends
		if token = boombox.mqttclient.Publish(data[1], 0, false, payload); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}

		count++
		if boombox.isCounter {
			fmt.Fprintf(os.Stderr, "\rPLAY #%d", count)
		}
	}
	if err = scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stderr, "\rEND  #%d\n", count)
}

func (boombox *boombox) close() {
	boombox.mqttclient.Disconnect(0)
	fmt.Fprintln(os.Stderr, "\rDISCONNECTED")
}

# MQTTBoombox
A versatile boombox for your MQTT broker.

## SYNOPSIS
```
  mqttboombox [--broker=<url>]
              [--binary] [--ff] [--counter]
              [subscribe-topic]...
```

## DESCRIPTION
mqttboombox listens to `subscribe-topic` (default to wildcard "#") and records
timing, topic and payload of the messages to STDOUT. Data read from STDIN will
be played/published back in the same order and timing. STDERR is used only for
infos.

mqttboombox can also record and play binary data.

## INSTALL
```
  go get github.com/eclipse/paho.mqtt.golang
  go get github.com/raffaelecassia/mqttboombox
  go build && go install
```

## OPTIONS
```
  --broker=<url>
      broker URL (default "tcp://localhost:1883").

  --binary
      threat messages payload as binary data. It will be encoded in base64.

  --ff
      fastforward, skip messages timing. Meaningful only in playback. It's
      useful when you want to send an entire file or when piping two
      mqttboombox.

  --counter
      print messages counter.
```

## EXAMPLES

Record "a/topic/#" and "an/another/topic" from mqtt1 to file
```
  mqttboombox --broker=tcp://mqtt1:1883 --counter "a/topic/#" "an/another/topic" > record1.mqtt
```

Playback messages from file to default broker on localhost
```
  mqttboombox --counter < record1.mqtt
```

Pipe messages from mqtt1 to mqtt2
```
  mqttboombox --broker=tcp://mqtt1:1883 --counter | mqttboombox --broker=tcp://mqtt2:1883 --ff
```

Pipe messages from mqtt1 to file and to another broker
```
  mqttboombox --broker=tcp://mqtt1:1883 | tee record2.mqtt | mqttboombox --ff
```

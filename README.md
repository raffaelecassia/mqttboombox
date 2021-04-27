# MQTTBoombox
A versatile boombox for your MQTT broker.

## SYNOPSIS
```
  mqttboombox [--broker=<url>]
              [--binary]  [--ff]  [--counter]
              [--clientid string]
              [--username string]  [--password string]
              [--alpn string]
              [--cafile string]  [--cert string --key string]
              [subscribe-topic]...
```

## DESCRIPTION
mqttboombox listens to `subscribe-topic`(s) (default to wildcard "#") and records
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
  --broker 
      broker URL (default "tcp://localhost:1883").
      to connect to a secure server use ssl:// prefix.

  --binary
      threat messages payload as binary data. It will be encoded in base64.

  --ff
      fastforward, skip messages timing. Meaningful only in playback. It's
      useful when you want to send an entire file or when piping two
      mqttboombox.

  --counter
      print messages counter to stderr.

  --clientid string
    	client id

  --username string
    	basic auth username
  --password string
    	basic auth password

  --alpn string
    	ALPN protocol name

  --cafile string
    	CA root certificate file (PEM format)
  --cert string
    	client certificate file (PEM format)
  --key string
    	client private key file (PEM format)
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

### AWS IOTCore

Connection with a custom authentication
```
mqttboombox \
  --broker ssl://XYZ-ats.iot.us-east-2.amazonaws.com:443 \
  --alpn mqtt \
  --clientid THING \
  --username THING \
  --password YYY
```

X.509 client certificate connection
```
mqttboombox \
  --broker ssl://XYZ-ats.iot.us-east-2.amazonaws.com:8883 \
  --clientid THING \
  --username THING \
  --cafile /path/to/AmazonRootCA1.pem \
  --cert /path/to/certificate.pem \
  --key /path/to/privatekey.pem
```

X.509 client certificate connection on port 443 with alpn
```
mqttboombox \
  --broker ssl://XYZ-ats.iot.us-east-2.amazonaws.com:443 \
  --alpn x-amzn-mqtt-ca \
  --clientid THING \
  --username THING \
  --cafile /path/to/AmazonRootCA1.pem \
  --cert /path/to/certificate.pem \
  --key /path/to/privatekey.pem
```

Reference: https://docs.aws.amazon.com/iot/latest/developerguide/protocols.html


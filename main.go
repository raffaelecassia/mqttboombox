package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"
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
	alpnproto := flag.String("alpn", "", "ALPN protocol")
	certfile := flag.String("cert", "", "client certificate PEM")
	keyfile := flag.String("key", "", "client private key PEM")
	cafile := flag.String("cafile", "", "CA root certificate PEM")
	flag.Parse()
	topics := flag.Args()

	bb = New(*brokerURL, topics, *isBinary, *isCounter, *isFastforward)

	if *clientID != "" {
		bb.options.SetClientID(*clientID)
	}

	// basic username/password
	if *username != "" {
		bb.options.SetUsername(*username)
	}
	if *password != "" {
		bb.options.SetPassword(*password)
	}

	// certs
	tlsConfig := tls.Config{}
	setTlsConfig := false
	if *certfile != "" && *keyfile != "" {
		// https://pkg.go.dev/crypto/tls#LoadX509KeyPair
		cert, err := tls.LoadX509KeyPair(*certfile, *keyfile)
		if err != nil {
			log.Fatal(err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
		setTlsConfig = true
	}
	if *cafile != "" {
		certPool := x509.NewCertPool()
		pemData, err := ioutil.ReadFile(*cafile)
		if err != nil {
			log.Fatal(err)
		}
		certPool.AppendCertsFromPEM(pemData)
		tlsConfig.RootCAs = certPool
		setTlsConfig = true
	}
	if *alpnproto != "" {
		tlsConfig.NextProtos = []string{*alpnproto}
		setTlsConfig = true
	}
	if setTlsConfig {
		bb.options.SetTLSConfig(&tlsConfig)
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

# Native Go MQTT Library

[![Codeship Status for FluuxIO/mqtt](https://app.codeship.com/projects/75c09d70-d43d-0135-b59a-12b6e6b26eee/status?branch=master)](https://app.codeship.com/projects/262977)
[![GoDoc](https://godoc.org/fluux.io/mqtt?status.svg)](https://godoc.org/fluux.io/mqtt) [![GoReportCard](https://goreportcard.com/badge/fluux.io/mqtt)](https://goreportcard.com/report/fluux.io/mqtt) [![codecov](https://codecov.io/gh/FluuxIO/mqtt/branch/master/graph/badge.svg)](https://codecov.io/gh/FluuxIO/mqtt)

Fluux MQTT is a MQTT v3.1.1 client library written in Go.

The library has been tested with the following MQTT servers:

- [ejabberd](https://www.process-one.net/en/ejabberd/)
- [fluux.io platform](https://fluux.io/)
- [Mosquitto](https://mosquitto.org/)

## Features

- MQTT v3.1.1, QOS 0
- Client manager to support auto-reconnect with exponential backoff.
- TLS Support

## Short term tasks

Implement support for QOS 1 and 2 (with storage backend interface and default backends).

## Running tests

You can launch unit tests with:

    go test ./...

## Testing with Fluux public MQTT server

We encourage you to experiment and test on a public Fluux test server. It is available on mqtt.fluux.io (on ports 1883
for cleartext and 8883 for TLS).

Here is example code for a simple client:

```
package main

import (
	"log"
	"time"

	"gosrc.io/mqtt"
)

func main() {
	client := mqtt.NewClient("tls://mqtt.fluux.io:8883")
	client.ClientID = "MQTT-Sub"
	log.Printf("Connecting on: %s\n", client.Address)

	messages := make(chan mqtt.Message)
	client.Messages = messages

	postConnect := func(c *mqtt.Client) {
		log.Println("Connected")
		name := "/mremond/test-topic-1"
		topic := mqtt.Topic{Name: name, QOS: 0}
		c.Subscribe(topic)
	}

	cm := mqtt.NewClientManager(client, postConnect)
	cm.Start()

	for m := range messages {
		log.Printf("Received message from MQTT server on topic %s: %+v\n", m.Topic, m.Payload)
	}
}
```

## Setting Mosquitto on OSX for testing

If you want to test Go MQTT library locally, you can install Mosquitto.

Mosquitto can be installed from homebrew:

```
brew install mosquitto
...
mosquitto has been installed with a default configuration file.
You can make changes to the configuration by editing:
    /usr/local/etc/mosquitto/mosquitto.conf

To have launchd start mosquitto at login:
  ln -sfv /usr/local/opt/mosquitto/*.plist ~/Library/LaunchAgents
Then to load mosquitto now:
  launchctl load ~/Library/LaunchAgents/homebrew.mxcl.mosquitto.plist
Or, if you don't want/need launchctl, you can just run:
  mosquitto -c /usr/local/etc/mosquitto/mosquitto.conf
```

Default config file can be customized in `/usr/local/etc/mosquitto/mosquitto.conf`.
However, default config file should be ok for testing

You can launch Mosquitto broker with command:

```
/usr/local/sbin/mosquitto -c /usr/local/etc/mosquitto/mosquitto.conf
```

The following command can be use to subscribe a client:

```
mosquitto_sub -v -t 'test/topic'
```

You can publish a payload payload on a topic with:

```
mosquitto_pub -t "test/topic" -m "message payload" -q 1
```

## Setting Mosquitto for testing on Windows 10

After you have install official Mosquitto build from main site, you can run the broker with command:

```
.\mosquitto.exe -v -c .\mosquitto.conf
```

You can subscribe with:

```
.\mosquitto_sub.exe -h 127.0.0.1 -v -t 'test/topic'
```

You can test publish with:

```
.\mosquitto_pub.exe -h 127.0.0.1 -t "test/topic" -m "message payload" -q 1
```

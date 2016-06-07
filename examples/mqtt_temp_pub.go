// +build darwin
package main

import (
	"log"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/processone/gomqtt/mqtt"
)

func main() {
	client := mqtt.New("localhost:1883", nil)
	client.ClientID = "mremond-osx"

	if err := client.Connect(); err != nil {
		log.Fatal("Connection error: ", err)
	}

	ticker := time.NewTicker(5 * time.Second)
	stop := make(chan bool)
	go publishLoop(client, ticker, stop)
	runtime.Goexit()
}

func publishLoop(client *mqtt.Client, ticker *time.Ticker, stop <-chan bool) {
	for done := false; !done; {
		select {
		case <-ticker.C:
			payload := make([]byte, 1, 1)
			payload[0] = getTemp()
			client.Publish(getTopic(client.ClientID), payload)
		case <-stop:
			done = true
			break
		}
	}
}

func getTopic(id string) string {
	return strings.Join([]string{id, "/cputemp"}, "")
}

func getTemp() byte {
	out, err := exec.Command("sysctl", "-n", "machdep.xcpm.cpu_thermal_level").Output()
	if err != nil {
		log.Println("Cannot read CPU temperature: ", err)
		return byte(0)
	}
	s := string(out)
	if temp, err := strconv.ParseInt(strings.Trim(s, "\n"), 10, 32); err != nil {
		return byte(temp)
	}
	return byte(0)
}

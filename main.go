package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/MichaelS11/go-ads"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stianeikeland/go-rpio/v4"
)

var (
	pump1 = rpio.Pin(9)
	pump2 = rpio.Pin(25)
	pump3 = rpio.Pin(11)
	pump4 = rpio.Pin(8)
)

func main() {
	err := ads.HostInit()
	if err != nil {
		log.Fatalln(err)
	}

	// create new ads with wanted busName and address. 
	ads1, err := ads.NewADS("I2C1", 0x48, "")
	if err != nil {
		log.Fatalln(err)
	}
	defer ads1.Close()

	ads1.SetConfigGain(ads.ConfigGain1)

	for {
		result, err := ads1.ReadRetry(5)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(result)
		time.Sleep(time.Second)
	}

	// init gpio
	if err := rpio.Open(); err != nil {
		log.Fatalln(err)
	}
	defer rpio.Close()

	// init pins
	pump1.Output()
	pump1.High()
	pump2.Output()
	pump2.High()
	pump3.Output()
	pump3.High()
	pump4.Output()
	pump4.High()

	// test pump 1
	pump1.Low()
	time.Sleep(time.Second)
	pump1.High()

	// test pump 2
	pump2.Low()
	time.Sleep(time.Second)
	pump2.High()

	// test pump 3
	pump3.Low()
	time.Sleep(time.Second)
	pump3.High()

	// test pump 4
	pump4.Low()
	time.Sleep(time.Second)
	pump4.High()

	// serve metrics til the end of days
	log.Println("listening on 0.0.0.0:5000")
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe("0.0.0.0:5000", nil)
}

package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/MichaelS11/go-ads"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stianeikeland/go-rpio/v4"
)

// calibrated to my Gikfun soil moisture sensor
// https://www.amazon.com/Gikfun-Capacitive-Corrosion-Resistant-Detection/dp/B07H3P1NRM
const (
	SoilWet = 8500
	SoilDry = 19000
)

type Seahorse struct {
	sync.Mutex

	sensor *ads.ADS
	pump   rpio.Pin

	soilMoisture prometheus.Gauge
	pumpUptime   prometheus.Counter
}

func NewSeahorse(sensor *ads.ADS, pump rpio.Pin) (*Seahorse, error) {
	// init and register prometheus metrics
	soilMoisture := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "seahorse_soil_moisture",
		Help: "Current soil moisture level",
	})
	prometheus.MustRegister(soilMoisture)

	pumpUptime := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "seahorse_pump_uptime",
		Help: "Total pump uptime in milliseconds",
	})
	prometheus.MustRegister(pumpUptime)

	seahorse := Seahorse{
		sensor: sensor,
		pump:   pump,

		soilMoisture:   soilMoisture,
		pumpUptime: pumpUptime,
	}
	return &seahorse, nil
}

func (s *Seahorse) ReadMoisture() (float64, error) {
	s.Lock()
	defer s.Unlock()

	// check the current moisture level
	result, err := s.sensor.ReadRetry(5)
	if err != nil {
		return 0, err
	}

	// clamp between 0 and 1
	moisture := float64(result-SoilWet) / float64(SoilDry-SoilWet)
	moisture = 1.0 - clamp(moisture, 0.0, 1.0)

	return moisture, nil
}

func (s *Seahorse) RunPump(duration time.Duration) {
	// run the pump for requested duration
	s.pump.Low()
	time.Sleep(duration)
	s.pump.High()
}

func (s *Seahorse) TrackMoisture(interval time.Duration) {
	c := time.Tick(interval)
	for {
		// check the current moisture level
		moisture, err := s.ReadMoisture()
		if err != nil {
			// if we see an error here, wait a bit and try again
			log.Println(err)
			time.Sleep(5 * time.Second)
			continue
		}

		// update metric
		s.soilMoisture.Set(moisture)

		// sleep til next loop
		<-c
	}
}

func (s *Seahorse) ControlLoop(interval time.Duration) {
	c := time.Tick(interval)
	for {
		// check the current moisture level
		moisture, err := s.ReadMoisture()
		if err != nil {
			// if we see an error here, wait a bit and try again
			log.Println(err)
			time.Sleep(5 * time.Second)
			continue
		}

		// if soil is dry, turn on the pump for a few seconds and update metric
		if moisture < 0.25 {
			s.RunPump(5 * time.Second)
			s.pumpUptime.Add(5000)
		}

		// sleep til next loop
		<-c
	}
}

func main() {
	// init gpio
	if err := rpio.Open(); err != nil {
		log.Fatalln(err)
	}
	defer rpio.Close()

	// init gpio pin(s)
	//	pump0 = 9
	//	pump1 = 25
	//	pump2 = 11
	//	pump3 = 8
	pump := rpio.Pin(9)
	pump.Output()
	pump.High()

	// setup Adafruit ADS1115 connection
	err := ads.HostInit()
	if err != nil {
		log.Fatalln(err)
	}

	// connect to addr 0x48 on bus 1
	sensor, err := ads.NewADS("I2C1", 0x48, "")
	if err != nil {
		log.Fatalln(err)
	}
	defer sensor.Close()

	// set gain and mode
	sensor.SetConfigGain(ads.ConfigGain1)
	sensor.SetConfigInputMultiplexer(ads.ConfigInputMultiplexerSingle0)

	// ready the seahorse!
	seahorse, err := NewSeahorse(sensor, pump)
	if err != nil {
		log.Fatalln(err)
	}

	// kick off moisture tracking
	go seahorse.TrackMoisture(15 * time.Second)

	// kick off control loop
	go seahorse.ControlLoop(15 * time.Minute)

	// serve metrics forever on the main goroutine
	log.Println("listening on 0.0.0.0:5000")
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe("0.0.0.0:5000", nil)
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

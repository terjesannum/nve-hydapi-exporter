package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	stationIds []string
	key        string
	stations   []Station
	interval   int
	maxAge     int
)

type Station struct {
	Data []struct {
		Id         string  `json:"stationId"`
		Name       string  `json:"stationName"`
		Masl       int     `json:"masl"`
		LakeName   string  `json:"LakeName"`
		RiverName  string  `json:"riverName"`
		Latitude   float64 `json:"latitude"`
		Longitude  float64 `json:"longitude"`
		SeriesList []struct {
			ParameterName string `json:"parameterName"`
			Parameter     int    `json:"parameter"`
			Unit          string `json:"unit"`
		} `json:"seriesList"`
	} `json:"data"`
	Measurements map[int]float64
	Valid        map[int]bool
}

type observations struct {
	Data []struct {
		Observations []struct {
			Time  time.Time `json:"time"`
			Value float64   `json:"value"`
		} `json:"observations"`
	} `json:"data"`
}

type nveCollector struct {
	stations         []Station
	stationInfo      *prometheus.Desc
	waterLevel       *prometheus.Desc
	airTemperature   *prometheus.Desc
	airHumidity      *prometheus.Desc
	waterTemperature *prometheus.Desc
	waterFlow        *prometheus.Desc
	windDirection    *prometheus.Desc
	windSpeed        *prometheus.Desc
}

func newNveCollector(stations []Station) *nveCollector {
	return &nveCollector{
		stations: stations,
		stationInfo: prometheus.NewDesc(
			"nve_station_info",
			"Station info",
			[]string{"station_id", "name", "lake", "river", "masl", "latitude", "longitude"},
			nil,
		),
		waterLevel: prometheus.NewDesc(
			"nve_station_water_level",
			"Waterlevel",
			[]string{"station_id"},
			nil,
		),
		airTemperature: prometheus.NewDesc(
			"nve_station_air_temperature",
			"Air temperature",
			[]string{"station_id"},
			nil,
		),
		airHumidity: prometheus.NewDesc(
			"nve_station_air_humidity",
			"Air humidity",
			[]string{"station_id"},
			nil,
		),
		waterTemperature: prometheus.NewDesc(
			"nve_station_water_temperature",
			"Water temperature",
			[]string{"station_id"},
			nil,
		),
		waterFlow: prometheus.NewDesc(
			"nve_station_water_flow",
			"Water flow",
			[]string{"station_id"},
			nil,
		),
		windDirection: prometheus.NewDesc(
			"nve_station_wind_direction",
			"Wind direction",
			[]string{"station_id"},
			nil,
		),
		windSpeed: prometheus.NewDesc(
			"nve_station_wind_speed",
			"Wind speed",
			[]string{"station_id"},
			nil,
		),
	}
}

func (collector *nveCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.stationInfo
	ch <- collector.waterLevel
	ch <- collector.airTemperature
	ch <- collector.airHumidity
	ch <- collector.waterTemperature
	ch <- collector.waterFlow
	ch <- collector.windDirection
	ch <- collector.windSpeed
}

func (collector *nveCollector) Collect(ch chan<- prometheus.Metric) {
	for _, station := range collector.stations {
		ch <- prometheus.MustNewConstMetric(
			collector.stationInfo,
			prometheus.GaugeValue,
			1,
			station.Data[0].Id,
			station.Data[0].Name,
			station.Data[0].LakeName,
			station.Data[0].RiverName,
			fmt.Sprintf("%d", station.Data[0].Masl),
			fmt.Sprintf("%f", station.Data[0].Latitude),
			fmt.Sprintf("%f", station.Data[0].Longitude),
		)
		for _, s := range station.Data[0].SeriesList {
			if station.Valid[s.Parameter] {
				if s.ParameterName == "Vannstand" {
					ch <- prometheus.MustNewConstMetric(
						collector.waterLevel,
						prometheus.GaugeValue,
						station.Measurements[s.Parameter],
						station.Data[0].Id,
					)
				} else if s.ParameterName == "Vannføring" {
					ch <- prometheus.MustNewConstMetric(
						collector.waterFlow,
						prometheus.GaugeValue,
						station.Measurements[s.Parameter],
						station.Data[0].Id,
					)
				} else if s.ParameterName == "Lufttemperatur" {
					ch <- prometheus.MustNewConstMetric(
						collector.airTemperature,
						prometheus.GaugeValue,
						station.Measurements[s.Parameter],
						station.Data[0].Id,
					)
				} else if s.ParameterName == "Relativ luftfuktighet" {
					ch <- prometheus.MustNewConstMetric(
						collector.airHumidity,
						prometheus.GaugeValue,
						station.Measurements[s.Parameter],
						station.Data[0].Id,
					)
				} else if s.ParameterName == "Vanntemperatur" {
					ch <- prometheus.MustNewConstMetric(
						collector.waterTemperature,
						prometheus.GaugeValue,
						station.Measurements[s.Parameter],
						station.Data[0].Id,
					)
				} else if s.ParameterName == "Vindretning" {
					ch <- prometheus.MustNewConstMetric(
						collector.windDirection,
						prometheus.GaugeValue,
						station.Measurements[s.Parameter],
						station.Data[0].Id,
					)
				} else if s.ParameterName == "Vindhastighet" {
					ch <- prometheus.MustNewConstMetric(
						collector.windSpeed,
						prometheus.GaugeValue,
						station.Measurements[s.Parameter],
						station.Data[0].Id,
					)
				}
			}
		}
	}
}

func exitOnError(msg string, err error) {
	if err != nil {
		log.Println(msg)
		log.Printf("Error: %v\n", err)
		log.Println("Exiting...")
		time.Sleep(10 * time.Second)
		os.Exit(1)
	}
}

func init() {
	var s string
	flag.StringVar(&key, "key", os.Getenv("NVE_API_KEY"), "NVE api key")
	flag.IntVar(&interval, "interval", 10, "Update interval (minutes)")
	flag.IntVar(&maxAge, "max-age", 24, "Maxium age of observation (hours) to be included")
	flag.StringVar(&s, "stations", os.Getenv("NVE_STATIONS"), "Comma separated list of station ids")
	flag.Parse()
	for _, station := range strings.Split(s, ",") {
		stationIds = append(stationIds, station)
	}
}

func getJson(c *http.Client, url string, target interface{}) {
	log.Printf("Getting json data from %s\n", url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	exitOnError("Create http request failed", err)
	req.Header.Set("X-API-Key", key)
	req.Header.Set("Accept", "application/json")
	res, err := c.Do(req)
	exitOnError("Http request failed", err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	exitOnError("Failed reading http response", err)
	err = json.Unmarshal([]byte(body), &target)
	exitOnError(fmt.Sprintf("Json error: %s\n", body), err)
}

func getStation(c *http.Client, id string) Station {
	station := new(Station)
	getJson(c, fmt.Sprintf("https://hydapi.nve.no/api/v1/Stations?StationId=%s", id), &station)
	log.Printf("Found station: %s\n", station.Data[0].Name)
	station.Measurements = make(map[int]float64)
	station.Valid = make(map[int]bool)
	return *station
}

func (station *Station) updateData(c *http.Client, parameter int) {
	obs := new(observations)
	getJson(c, fmt.Sprintf("https://hydapi.nve.no/api/v1/Observations?StationId=%s&ResolutionTime=0&Parameter=%d", station.Data[0].Id, parameter), &obs)
	timeDiff := time.Now().Sub(obs.Data[0].Observations[0].Time)
	station.Measurements[parameter] = obs.Data[0].Observations[0].Value
	if timeDiff.Hours() > float64(maxAge) {
		// flag observation as invalid if older than maxAge
		station.Valid[parameter] = false
		log.Printf("Too old observation %d for %s: %s\n", parameter, station.Data[0].Id, obs.Data[0].Observations[0].Time)
	} else {
		station.Valid[parameter] = true
	}
	if station.Valid[parameter] {
		if parameter == 1000 && station.Data[0].Masl > 10 && obs.Data[0].Observations[0].Value <= 0 {
			// Some stations seems to sometimes report 0 as waterlevel,
			// flag as invalid for stations placed more than 10 meters above sea level
			station.Valid[parameter] = false
			log.Printf("Invalid observation %d for %s: %f\n", parameter, station.Data[0].Id, obs.Data[0].Observations[0].Value)
		} else if parameter == 1001 && obs.Data[0].Observations[0].Value <= 0 {
			// Drop water flow observations <= 0
			station.Valid[parameter] = false
			log.Printf("Invalid observation %d for %s: %f\n", parameter, station.Data[0].Id, obs.Data[0].Observations[0].Value)
		}
	}
}

func (station *Station) startCollector(ctx context.Context, client *http.Client, parameter int) {
	station.updateData(client, parameter)
	ticker := time.NewTicker(time.Duration(interval) * time.Minute)
	quit := make(chan struct{})
	go func(client *http.Client, station Station, parameter int) {
		for {
			select {
			case <-ticker.C:
				station.updateData(client, parameter)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}(client, *station, parameter)
}

func main() {
	ctx := context.Background()
	client := http.Client{
		Timeout: time.Second * 10,
	}
	for _, id := range stationIds {
		station := getStation(&client, id)
		stations = append(stations, station)
		time.Sleep(time.Second) // avoid hitting rate limit of 5 req/s
	}
	for _, station := range stations {
		for _, serie := range station.Data[0].SeriesList {
			if serie.ParameterName == "Vannstand" {
				station.startCollector(ctx, &client, serie.Parameter)
			} else if serie.ParameterName == "Vannføring" {
				station.startCollector(ctx, &client, serie.Parameter)
			} else if serie.ParameterName == "Lufttemperatur" {
				station.startCollector(ctx, &client, serie.Parameter)
			} else if serie.ParameterName == "Relativ luftfuktighet" {
				station.startCollector(ctx, &client, serie.Parameter)
			} else if serie.ParameterName == "Vanntemperatur" {
				station.startCollector(ctx, &client, serie.Parameter)
			} else if serie.ParameterName == "Vindretning" {
				station.startCollector(ctx, &client, serie.Parameter)
			} else if serie.ParameterName == "Vindhastighet" {
				station.startCollector(ctx, &client, serie.Parameter)
			}
			time.Sleep(time.Second) // avoid hitting rate limit of 5 req/s
		}
	}
	collector := newNveCollector(stations)
	prometheus.MustRegister(collector)
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":8080", nil)
	exitOnError("Http listener failed", err)
}

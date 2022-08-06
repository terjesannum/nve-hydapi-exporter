# Get NVE hydrological data in prometheus

This prometheus exporter will get data from [NVE's HydAPI](https://hydapi.nve.no/) and make them available as prometheus metrics.

## Running

Docker image is available on [ghcr.io](https://github.com/terjesannum/nve-hydapi-exporter/pkgs/container/nve-hydapi-exporter).

```sh
docker run -d -p 8080:8080 --restart always ghcr.io/terjesannum/nve-hydapi-exporter:3 --key ... --stations 12.215.0,12.611.0 --interval 10 --max-age 24
```

Environment variables `NVE_API_KEY` and `NVE_STATIONS` can also be used for the options.

Use [NVE's Sildre](https://sildre.nve.no/) to find the ids for the stations you wish to monitor.

## Metrics

```
# HELP nve_station_info Station info
# TYPE nve_station_info gauge
nve_station_info{lake="",latitude="60.891720",longitude="8.332810",masl="900",name="Storeskar",river="Hemsil",station_id="12.215.0"} 1
nve_station_info{lake="",latitude="60.695260",longitude="9.018590",masl="211",name="Liaåni",river="Hallingdalsvassdrage",station_id="12.611.0"} 1

# HELP nve_station_water_flow Water flow
# TYPE nve_station_water_flow gauge
nve_station_water_flow{station_id="12.215.0"} 1.709347

# HELP nve_station_water_level Waterlevel
# TYPE nve_station_water_level gauge
nve_station_water_level{station_id="12.215.0"} 1.082
nve_station_water_level{station_id="12.611.0"} 208.172

# HELP nve_station_air_temperature Air temperature
# TYPE nve_station_air_temperature gauge
nve_station_air_temperature{station_id="12.215.0"} 11

# HELP nve_station_water_temperature Water temperature
# TYPE nve_station_water_temperature gauge
nve_station_water_temperature{station_id="12.611.0"} 14
```

Note that what metrics that are available from different stations varies.
Currently **water level**, **water flow**, **water temperature** and **air temperature** are supported.

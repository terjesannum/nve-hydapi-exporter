# Get NVE hydrological data in prometheus

This prometheus exporter will get data from [NVE's HydAPI](https://hydapi.nve.no/) and make them available as prometheus metrics.

## Running

Docker image is available on [ghcr.io](https://github.com/terjesannum/nve-hydapi-exporter/pkgs/container/nve-hydapi-exporter).

```sh
docker run -d -p 8080:8080 --restart always ghcr.io/terjesannum/nve-hydapi-exporter:5 --key ... --stations 12.215.0,12.611.0 --interval 10 --max-age 24
```

Environment variables `NVE_API_KEY` and `NVE_STATIONS` can also be used for the options.

Use [NVE's Sildre](https://sildre.nve.no/) to find the ids for the stations you wish to monitor.

## Metrics

```
# HELP nve_station_info Station info
# TYPE nve_station_info gauge
nve_station_info{lake="",latitude="60.891720",longitude="8.332810",masl="900",name="Storeskar",river="Hemsil",station_id="12.215.0"} 1
nve_station_info{lake="",latitude="60.695260",longitude="9.018590",masl="211",name="Liaåni",river="Hallingdalsvassdrage",station_id="12.611.0"} 1
nve_station_info{lake="",latitude="60.106720",longitude="10.683940",masl="372",name="Hakkloa, 0.5 km sør for Ø. Hakloa gård",river="Nordmarkvassdraget",station_id="6.24.4"} 1

# HELP nve_station_air_humidity Air humidity
# TYPE nve_station_air_humidity gauge
nve_station_air_humidity{station_id="6.24.4"} 95.23

# HELP nve_station_air_temperature Air temperature
# TYPE nve_station_air_temperature gauge
nve_station_air_temperature{station_id="12.215.0"} 11
nve_station_air_temperature{station_id="6.24.4"} 15.55

# HELP nve_station_water_flow Water flow
# TYPE nve_station_water_flow gauge
nve_station_water_flow{station_id="12.215.0"} 1.709347

# HELP nve_station_water_level Waterlevel
# TYPE nve_station_water_level gauge
nve_station_water_level{station_id="12.215.0"} 1.082
nve_station_water_level{station_id="12.611.0"} 208.172

# HELP nve_station_water_temperature Water temperature
# TYPE nve_station_water_temperature gauge
nve_station_water_temperature{station_id="12.611.0"} 14

# HELP nve_station_wind_direction Wind direction
# TYPE nve_station_wind_direction gauge
nve_station_wind_direction{station_id="6.24.4"} 52

# HELP nve_station_wind_speed Wind speed
# TYPE nve_station_wind_speed gauge
nve_station_wind_speed{station_id="6.24.4"} 1.2
```

Note that what metrics that are available from different stations varies.
Currently **water level**, **water flow**, **water temperature**, **air humidity**, **air temperature**, **wind direction** and **wind speed** are supported.

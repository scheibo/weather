package weather

import (
	"fmt"
	"time"

	"github.com/scheibo/darksky"
	"github.com/scheibo/geo"
)

var ICONS = []string{
	"clear-day", "clear-night", "rain", "snow", "sleet", "wind",
	"fog", "cloudy", "partly-cloudy-day", "partly-cloudy-night",
}

type darkSkyProvider struct {
	client *darksky.Client
	loc    *time.Location
}

const PIRATE_WEATHER = "https://api.pirateweather.net/forecast"

func newDarkSkyProvider(key string, loc *time.Location) *darkSkyProvider {
	client := darksky.NewClient(key)
	client.BaseURL = PIRATE_WEATHER
	return &darkSkyProvider{client: client, loc: loc}
}

var darkSkyCurrentArguments = darksky.Arguments{"excludes": "minutely,hourly,alerts,flags", "units": "si"}
var darkSkyForecastArguments = darksky.Arguments{"excludes": "minutely,alerts,flags", "extend": "hourly", "units": "si"}
var darkSkyHistoryArguments = darkSkyCurrentArguments

func (w *darkSkyProvider) current(ll geo.LatLng) (*Conditions, error) {
	f, err := w.client.GetForecast(geo.Coordinate(ll.Lat), geo.Coordinate(ll.Lng), darkSkyCurrentArguments)
	if err != nil {
		return nil, err
	}
	if len(f.Daily.Data) < 1 {
		return nil, fmt.Errorf("missing daily data")
	}
	return DarkSkyToConditions(f.Currently, &f.Daily.Data[0], w.loc), nil
}

func (w *darkSkyProvider) forecast(ll geo.LatLng) (*Forecast, error) {
	f, err := w.client.GetForecast(geo.Coordinate(ll.Lat), geo.Coordinate(ll.Lng), darkSkyForecastArguments)
	if err != nil {
		return nil, err
	}

	days := make(map[int]*darksky.DataPoint)
	for _, d := range f.Daily.Data {
		dp := d
		days[d.Time.Time.In(w.loc).YearDay()] = &dp
	}

	forecast := Forecast{}
	for _, h := range f.Hourly.Data {
		d, _ := days[h.Time.Time.In(w.loc).YearDay()]
		forecast.Hourly = append(forecast.Hourly, DarkSkyToConditions(&h, d, w.loc))
	}

	return &forecast, nil
}

func (w *darkSkyProvider) history(ll geo.LatLng, t time.Time) (*Conditions, error) {
	f, err := w.client.GetTimeMachineForecast(geo.Coordinate(ll.Lat), geo.Coordinate(ll.Lng), t, darkSkyHistoryArguments)
	if err != nil {
		return nil, err
	}
	if len(f.Daily.Data) < 1 {
		return nil, fmt.Errorf("missing daily data")
	}
	return DarkSkyToConditions(f.Currently, &f.Daily.Data[0], w.loc), nil
}

func DarkSkyToConditions(h *darksky.DataPoint, d *darksky.DataPoint, loc *time.Location) *Conditions {
	// BUG: These values are marked 'optional' by DarkSky, so it could return
	// nothing for one of these and we would mistake it for 0 (which is otherwise
	// a completely valid data point).
	c := Conditions{
		Icon:                h.Icon,
		Time:                h.Time.Time.In(loc),
		Temperature:         h.Temperature,
		ApparentTemperature: h.ApparentTemperature,
		Humidity:            h.Humidity,
		PrecipProbability:   h.PrecipProbability,
		PrecipIntensity:     h.PrecipIntensity,
		PrecipType:          h.PrecipType,
		AirPressure:         h.Pressure,
		AirDensity:          rho(h.Temperature, h.Pressure, h.DewPoint),
		CloudCover:          h.CloudCover,
		UVIndex:             h.UVIndex,
		WindSpeed:           h.WindSpeed,
		WindGust:            h.WindGust,
		WindBearing:         h.WindBearing,
	}
	if d != nil {
		c.SunriseTime = time.Unix(int64(d.SunriseTime), 0).In(loc)
		c.SunsetTime = time.Unix(int64(d.SunsetTime), 0).In(loc)
	}
	return &c
}

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

type DarkSkyProvider struct {
	client *darksky.Client
	loc    *time.Location
}

func NewDarkSkyProvider(key string, tz ...*time.Location) *DarkSkyProvider {
	loc := time.UTC
	if len(tz) > 0 && tz[0] != nil {
		loc = tz[0]
	}
	return &DarkSkyProvider{client: darksky.NewClient(key), loc: loc}
}

var DarkSkyCurrentArguments = darksky.Arguments{"excludes": "minutely,hourly,alerts,flags", "units": "si"}
var DarkSkyForecastArguments = darksky.Arguments{"excludes": "minutely,alerts,flags", "extend": "hourly", "units": "si"}
var DarkSkyHistoryArguments = DarkSkyCurrentArguments

func (w *DarkSkyProvider) Current(ll geo.LatLng) (*Conditions, error) {
	f, err := w.client.GetForecast(geo.Coordinate(ll.Lat), geo.Coordinate(ll.Lng), DarkSkyCurrentArguments)
	if err != nil {
		return nil, err
	}
	if len(f.Daily.Data) < 1 {
		return nil, fmt.Errorf("missing daily data")
	}
	return w.ToConditions(f.Currently, &f.Daily.Data[0]), nil
}

func (w *DarkSkyProvider) Forecast(ll geo.LatLng) (*Forecast, error) {
	f, err := w.client.GetForecast(geo.Coordinate(ll.Lat), geo.Coordinate(ll.Lng), DarkSkyForecastArguments)
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
		forecast.Hourly = append(forecast.Hourly, w.ToConditions(&h, d))
	}

	return &forecast, nil
}

func (w *DarkSkyProvider) History(ll geo.LatLng, t time.Time) (*Conditions, error) {
	f, err := w.client.GetTimeMachineForecast(geo.Coordinate(ll.Lat), geo.Coordinate(ll.Lng), t, DarkSkyHistoryArguments)
	if err != nil {
		return nil, err
	}
	if len(f.Daily.Data) < 1 {
		return nil, fmt.Errorf("missing daily data")
	}
	return w.ToConditions(f.Currently, &f.Daily.Data[0]), nil
}

func (w *DarkSkyProvider) ToConditions(h *darksky.DataPoint, d *darksky.DataPoint) *Conditions {
	// BUG: These values are marked 'optional' by DarkSky, so it could return
	// nothing for one of these and we would mistake it for 0 (which is otherwise
	// a completely valid data point).
	c := Conditions{
		Icon:                h.Icon,
		Time:                h.Time.Time.In(w.loc),
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
		c.SunriseTime = time.Unix(int64(d.SunriseTime), 0).In(w.loc)
		c.SunsetTime = time.Unix(int64(d.SunsetTime), 0).In(w.loc)
	}
	return &c
}

package weather

import (
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
}

func newDarkSkyProvider(key string) *darkSkyProvider {
	return &darkSkyProvider{client: darksky.NewClient(key)}
}

var darkSkyCurrentArguments = darksky.Arguments{"excludes": "minutely,hourly,daily,alerts,flags", "units": "si"}
var darkSkyForecastArguments = darksky.Arguments{"excludes": "minutely,alerts,flags", "extend": "hourly", "units": "si"}
var darkSkyHistoryArguments = darkSkyCurrentArguments

func (w *darkSkyProvider) current(ll geo.LatLng) (*Conditions, error) {
	f, err := w.client.GetForecast(geo.Coordinate(ll.Lat), geo.Coordinate(ll.Lng), darkSkyCurrentArguments)
	if err != nil {
		return nil, err
	}
	return w.toConditions(f.Currently), nil
}

func (w *darkSkyProvider) forecast(ll geo.LatLng) (*Forecast, error) {
	f, err := w.client.GetForecast(geo.Coordinate(ll.Lat), geo.Coordinate(ll.Lng), darkSkyForecastArguments)
	if err != nil {
		return nil, err
	}

	forecast := Forecast{}

	for _, h := range f.Hourly.Data {
		forecast.Hourly = append(forecast.Hourly, w.toConditions(&h))
	}

	return &forecast, nil
}

func (w *darkSkyProvider) history(ll geo.LatLng, t time.Time) (*Conditions, error) {
	f, err := w.client.GetTimeMachineForecast(geo.Coordinate(ll.Lat), geo.Coordinate(ll.Lng), t, darkSkyHistoryArguments)
	if err != nil {
		return nil, err
	}
	return w.toConditions(f.Currently), nil
}

func (w *darkSkyProvider) toConditions(dp *darksky.DataPoint) *Conditions {
	// BUG: These values are marked 'optional' by DarkSky, so it could return
	// nothing for one of these and we would mistake it for 0 (which is otherwise
	// a completely valid data point).
	return &Conditions{
		Icon:                dp.Icon,
		Time:                dp.Time.Time,
		Temperature:         dp.Temperature,
		ApparentTemperature: dp.ApparentTemperature,
		Humidity:            dp.Humidity,
		PrecipProbability:   dp.PrecipProbability,
		PrecipIntensity:     dp.PrecipIntensity,
		PrecipType:          dp.PrecipType,
		AirPressure:         dp.Pressure,
		AirDensity:          rho(dp.Temperature, dp.Pressure, dp.DewPoint),
		CloudCover:          dp.CloudCover,
		UVIndex:             dp.UVIndex,
		WindSpeed:           dp.WindSpeed,
		WindGust:            dp.WindGust,
		WindBearing:         dp.WindBearing,
	}
}

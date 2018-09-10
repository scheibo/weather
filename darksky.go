package weather

import (
	"fmt"
	"time"

	"github.com/adlio/darksky"
	"github.com/scheibo/geo"
)

type darkSkyProvider struct {
	client *darksky.Client
}

func newDarkSkyProvider(key string) *darkSkyProvider {
	return &darkSkyProvider{client: darksky.NewClient(key)}
}

var darkSkyCurrentArguments = darksky.Arguments{"excludes": "minutely,hourly,daily,alerts,flags", "units": "si"}
var darkSkyForecastArguments = darksky.Arguments{"excludes": "minutely,alerts,flags", "units": "si"}
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
	forecast.Currently = w.toConditions(f.Currently)

	if len(f.Hourly.Data) < 24 {
		return nil, fmt.Errorf("not enough hours returned in forecast")
	}

	for i := 0; i < 24; i++ {
		forecast.Hourly = append(forecast.Hourly, w.toConditions(&f.Hourly.Data[i]))
	}

	if len(f.Daily.Data) < 7 {
		return nil, fmt.Errorf("not enough days returned in forecast")
	}

	for i := 0; i < 7; i++ {
		forecast.Daily = append(forecast.Daily, w.toConditions(&f.Daily.Data[i]))
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
		Time:              dp.Time.Time,
		Temperature:       dp.Temperature,
		Humidity:          dp.Humidity,
		PrecipProbability: dp.PrecipProbability,
		PrecipIntensity:   dp.PrecipIntensity,
		AirPressure:       dp.Pressure,
		AirDensity:        rho(dp.Temperature, dp.Pressure, dp.DewPoint),
		WindSpeed:         dp.WindSpeed,
		WindBearing:       dp.WindBearing,
	}
}

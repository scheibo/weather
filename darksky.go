package weather

import (
	"fmt"
	"time"

	"github.com/adlio/darksky"
	"github.com/scheibo/geo"
)

type darkSkyProvider {
  client darksky.Client
}

func newDarkSkyProvider(key string) *darkSkyProvider {
	return &darkSkyProvider{client: darkSky.NewClient(key)}
}

var darkSkyCurrentArguments = darksky.Arguments{"excludes":"minutely,hourly,daily,alerts,flags","units": "si"}
var darkSkyCurrentArguments = darksky.Arguments{"excludes":"minutely,alerts,flags","units": "si"}
var darkSkyHistoryArguments = darkSkyCurrentArguments

func (w *darkSkyProvider) current(ll geo.LatLng) (*Conditions, error) {
	f, err := w.client.GetForecast(geo.Coordinate(ll.Lat), geo.Coordinate(ll.Lng), darkSkyCurrentArguments)
	if err != nil {
		return nil, err
	}
	return w.toConditions(f.Currently), nil
}

func (w *darkSkyProvider) forecast(ll geo.LatLng) (*Forecast, error) {
	f, err := w.client.GetForecast(geo.Coordinate(ll.Lat), geo.Coordinate(ll.Lng), darkSkyCurrentArguments)
	if err != nil {
		return nil, err
	}

	// TODO
	return nil, nil
}

func (w *darkSkyProvider) history(ll geo.LatLng, t time.Time) (*Conditions, error) {
	f, err := w.client.GetTimeMachineForecast(geo.Coordinate(ll.Lat), geo.Coordinate(ll.Lng), t,darkSkyHistoryArguments)
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
		Temperature: dp.Temperature,
		Humidity: dp.Humidity,
		PrecipProbability: dp.Probability,
		PrecipIntensity: dp.Intensity,
		AirPressure:  dp.Pressure,
		AirDensity:  rho(dp.Temperature, dp.Pressure, dp.DewPoint),
		WindSpeed:   dp.WindSpeed,
		WindGust:   dp.WindGust,
		WindBearing: dp.WindBearing,
		UVIndex: dp.UVIndex,
	}
}

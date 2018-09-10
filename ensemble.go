package weather

import (
	"fmt"
	"time"

	"github.com/scheibo/geo"
)

type keyWeight struct {
	key string
	weight float64
}

type options struct {
  ds, wu, noaa, owm	keyWeight
}

func adjustWeights(o *options) *options {
	keys := 0
	weights := 0.0

	keys += accumIfSet(o.ds, &weights)
	keys += accumIfSet(o.wu, &weights)
	keys += accumIfSet(o.noaa, &weights)
	keys += accumIfSet(o.owm, &weights)

	if weights == 1.0 {
		return o
	}

	if weights > 1.0 {
		mod := 1.0/weights
		modIfSet(o.ds.key, &o.ds.weight, mod)
		modIfSet(o.wu.key, &o.wu.weight, mod)
		modIfSet(o.noaa.key, &o.noaa.weight, mod)
		modIfSet(o.owm.key, &o.owm.weight, mod)
		return o
	}

	// weights < 1.0
	rem := 1.0 - weights
	w := rem/keys

	weightIfUnset(o.ds.key, &o.ds.weight, w)
	weightIfUnset(o.wu.key, &o.wu.weight, w)
	weightIfUnset(o.noaa.key, &o.noaa.weight, w)
	weightIfUnset(o.owm.key, &o.owm.weight, w)

	return o
}

func accumIfSet(kw keyWeight, weights *float64) int {
	if kw.key == "" {
		return 0
	}
	*weights += kw.weight
	return 1
}

func modIfSet(key string, weight *float64, mod float64) {
	if key != "" {
		*weight *= mod
	}
}

func weightIfUnset(key string, weight *float64, val *float64) {
	if key != "" && *weight == 0 {
		*weight = val
	}
}

type providerWeight struct {
	provider provider
	weight float64
}

type ensembleProvider {
  providers []*providerWeight
}

func configured(kw keyWeight) bool {
	return kw.key != "" && kw.weight != 0
}

func newEnsembleProvider(o *options) (*provider, error) {
	var providers []*providerWeights
	if configured(o.ds) {
		p := &providerWeights{
			provider: newDarkSkyProvider(options.ds.key),
			weight: options.ds.weight,
		}
	}
	// TODO(kjs): add other providers

	if len(providers) == 0 {
		return nil, fmt.Errorf("not configured with any weather providers")
	}

	return providers, nil
}

type ensembleProvider {
  providers []*providerWeight
}

func configured(kw keyWeight) bool {
	return kw.key != "" && kw.weight != 0
}

func newEnsembleProvider(o *options) (*provider, error) {
	var providers []*providerWeights
	if configured(o.ds) {
		p := &providerWeights{
			provider: newDarkSkyProvider(options.ds.key),
			weight: options.ds.weight,
		}
	}
	// TODO(kjs): add other providers

	if len(providers) == 0 {
		return nil, fmt.Errorf("not configured with any weather providers")
	}

	return providers, nil
}

type conditionsWeight struct {
	conditions *Conditions
	weight float64
}

func (w *ensembleProvider) current(ll geo.LatLng) (*Conditions, error) {
	return get(ll, time.Now(), func(p *provider, ll geo.LatLng, t time.Time) (*Conditions, error){
		return p.current(ll)
	})
}

func (w *ensembleProvider) history(ll geo.LatLng, t time.Time) (*Conditions, error) {
	return get(ll, time.Now(), func(p *provider, ll geo.LatLng, t time.Time) (*Conditions, error){
		return p.history(ll, t)
	})
}

func (w *ensembleProvider) get(ll geo.LatLng, t time.Time, call func(*provider, geo.LatLng, time.Time)(*Condition, error)) (*Condition, error) {
  conds := make(chan *conditionWeight, len(w))
	errs := make(chan error, len(w))

	for _, pw := range w.providers {
		go func(p provider) {
			c, err := call(p, ll, t)
			if err != nil {
				errs <- err
				return
			}
			conds <- &conditionWeight{condition: c, weight: pw.weight}
		}(pw.provider)
	}

	var conditions []*conditionWeight
	var errors []error

	for i := 0; i < len(w.providers); i++ {
		select {
		case cond := <-conds:
			conditions = append(conditions, cond)
		case err := <-errs:
			errors = append(errors, err)
		}
	}

	if len(conditions) > 0 {
		return w.weightCondition(conditions), nil
	}
	return nil, fmt.Errorf("errors: %+v", errors)
}

type forecastWeight struct {
	forecast *Forecast
	weight float64
}

func (w *ensembleProvider) forecast(ll geo.LatLng) (*Forecast, error) {
  fores := make(chan *forecastWeight, len(w))
	errs := make(chan error, len(w))

	for _, pw := range w.providers {
		go func(p provider) {
			f, err := p.forecast(ll)
			if err != nil {
				errs <- err
				return
			}
			fors <- &forecastWeight{forecast: f, weight: pw.weight}
		}(pw.provider)
	}

	var forecasts []*forecastWeight
	var errors []error

	for i := 0; i < len(w.providers); i++ {
		select {
		case fore := <-fores:
			forecasts = append(forecasts, fores)
		case err := <-errs:
			errors = append(errors, err)
		}
	}

	if len(forecasts) > 0 {
		return w.weightForecasts(forecasts), nil
	}
	return nil, fmt.Errorf("errors: %+v", errors)
}

func (w *ensembleProvider) weightConditions(conditions []*conditionWeight) (*Conditions, error) {
	c := Conditions{}

	weights := 0.0
	for _, cw := range conditions {
		weights += cw.weight
	}

	for _, cw := range conditions {
		w := cw.weight/weights
		c.Temperature += cw.conditions.Temperature * w
		c.Humidity += cw.conditions.Humidity * w
		c.PrecipProbability += cw.conditions.PrecipProbability * w
		c.PrecipIntensity += cw.conditions.PrecipIntensity * w
		c.AirPressure += cw.conditions.AirPressure * w
		c.AirDensity += cw.conditions.AirDensity * w
		c.WindSpeed += cw.conditions.WindSpeed * w
		c.WindGust += cw.conditions.WindGust * w
		c.WindBearing += cw.conditions.WindBearing * w // TODO average angles...
		c.UVIndex += cw.conditions.UVIndex * w
	}
}

func (w *ensembleProvider) weightForecast(forecasts []*forecastWeight) (*Forecast, error) {
	var hourlies [][]*Conditions
	var dailies [][]*Conditions

	for _, forecast := range forecasts {
		hourlies = append(hourlies, forecast.Hourly)
		dailies = append(dailies, forecast.Daily)
	}

	// TODO
	var hourly []*Condition
	for i = 0; i < len(hourlies); i++ {
		var conditions []*conditionWeights
		for j = 0; i < len(hourlies[i]); j++ { //hour
			con
		}
	}

	for _, hourly := range hourlies {
	}
}

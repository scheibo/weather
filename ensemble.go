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

func (w *ensembleProvider) current(ll geo.LatLng) (*Conditions, error) {
	return w.providers[0].current(ll) // TODO
}

func (w *ensembleProvider) forecast(ll geo.LatLng) (*Forecast, error) {
	return w.providers[0].forecast(ll) // TODO
}

func (w *ensembleProvider) history(ll geo.LatLng, t time.Time) (*Conditions, error) {
	return w.providers[0].history(ll) // TODO
}

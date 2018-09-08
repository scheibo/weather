package wtwi

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/adlio/darksky"
)

type LatLng struct {
	Lat, Lng float64
}

func (ll LatLng) String() string {
	return fmt.Sprintf("%s,%s", ll.Latitude(), ll.Longitude())
}

func (ll *LatLng) Latitude() string {
	return fmt.Sprintf("%.6f", ll.Lat)
}

func (ll *LatLng) Longitude() string {
	return fmt.Sprintf("%.6f", ll.Lng)
}

func ParseLatLng(s string) (LatLng, error) {
	sp := strings.Split(s, ",")
	if len(sp) != 2 {
		return LatLng{}, fmt.Errorf("expected 'latitude,longitude' pair")
	}

	lat, err := strconv.ParseFloat(strings.TrimSpace(sp[0]), 64)
	if err != nil {
		return LatLng{}, err
	}

	lng, err := strconv.ParseFloat(strings.TrimSpace(sp[1]), 64)
	if err != nil {
		return LatLng{}, err
	}

	return LatLng{Lat: lat, Lng: lng}, nil
}

type Weather struct {
	Temperature float64 `json:"temperature,omitempty"`
	AirDensity  float64 `json:"airDensity,omitempty"`
	WindSpeed   float64 `json:"windSpeed,omitempty"`
	WindBearing float64 `json:"windBearing,omitempty"`
}

func (w Weather) String() string {
	return fmt.Sprintf(
		"temp: %.1f °C, wind: %.1f km/h %s, density: %.3f kg/m³",
		round(w.Temperature, 0.1),
		round(w.WindSpeed*3600.0/1000.0, 0.1),
		bearingString(w.WindBearing),
		round(w.AirDensity, 0.001))
}

func Get(ll LatLng, t time.Time, keys ...string) (*Weather, error) {
	key := os.Getenv("DARKSKY_APIKEY")
	if len(keys) > 0 && keys[0] != "" {
		key = keys[0]
	}
	if key == "" {
		return nil, fmt.Errorf("must provide a DarkSky API key")
	}

	client := darksky.NewClient(key)
	f, err := client.GetTimeMachineForecast(ll.Latitude(), ll.Longitude(), t, darksky.Arguments{"units": "si"})
	if err != nil {
		return nil, err
	}

	// BUG: These values are marked 'optional' by DarkSky, so it could return
	// nothing for one of these and we would mistake it for 0 (which is otherwise
	// a completely valid data point).
	T := f.Currently.Temperature
	p := f.Currently.Pressure
	dp := f.Currently.DewPoint
	ws := f.Currently.WindSpeed
	wb := f.Currently.WindBearing

	println(bearingString(-1), bearingString(11.25), bearingString(181))

	return &Weather{
		Temperature: T,
		AirDensity:  rho(T, p, dp),
		WindSpeed:   ws,
		WindBearing: wb,
	}, nil
}

func rho(t, p, dp float64) float64 {
	const Rd = 287.0531 // specific gas constant for dry air in J(kg*K)
	const Rv = 461.4964 // specific gas constant for water vapor in J(kg*K)
	const K = 273.15    // the value of Kelvin corresponding to 0 Celsius.

	// Herman Wobus constants
	const c0 = 0.99999683
	const c1 = -0.90826951E-02
	const c2 = 0.78736169E-04
	const c3 = -0.61117958E-06
	const c4 = 0.43884187E-08
	const c5 = -0.29883885E-10
	const c6 = 0.21874425E-12
	const c7 = -0.17892321E-14
	const c8 = 0.11112018E-16
	const c9 = -0.30994571E-19

	x := c0 + dp*(c1+dp*(c2+dp*(c3+dp*(c4+dp*(c5+dp*(c6+dp*(c7+dp*(c8+dp*(c9)))))))))
	pv := 6.1078 / (math.Pow(x, 8))

	return 100 * (((p - pv) / (Rd * (t + K))) +
		(pv / (Rv * (t + K))))
}

func bearingString(wb float64) string {
	var COMPASS = []string{
		"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE",
		"S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW",
	}

	nwb := normalizeBearing(wb)
	index := int(math.Mod((nwb+11.25)/22.5, 16))
	dir := COMPASS[index]

	return fmt.Sprintf("%s (%.1f°)", dir, round(nwb, 0.5))
}

func round(x, unit float64) float64 {
	return math.Round(x/unit) * unit
}

func normalizeBearing(d float64) float64 {
	return d + math.Ceil(-d/360)*360
}

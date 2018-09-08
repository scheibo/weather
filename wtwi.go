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
	Lat string
	Lng string
}

func (ll LatLng) String() string {
	return fmt.Sprintf("%s,s", ll.Lng, ll.Lng)
}

func ParseLatLng(s string) (LatLng, error) {
	sp := strings.Split(s, ",")
	if len(sp) != 2 {
		return LatLng{}, fmt.Errorf("expected 'latitude,longitude' pair")
	}

	// NOTE: We want to ensure what we parsed are valid floats, but we don't
	// store the coordinates as floats to avoid precision issues.
	lat := strings.TrimSpace(sp[0])
	_, err := strconv.ParseFloat(lat, 64)
	if err != nil {
		return LatLng{}, err
	}

	lng := strings.TrimSpace(sp[1])
	_, err = strconv.ParseFloat(lng, 64)
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
		"temp:%.1f°C, wind: %.1fkm/h %s, density:%.3fkg/m³",
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
	f, err := client.GetTimeMachineForecast(ll.Lat, ll.Lng, t, darksky.Arguments{"units": "si"})
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
	var dir string

	nwb := normalizeBearing(wb)
	if nwb < 11.25 {
		dir = "N"
	} else if nwb >= 11.25 && nwb < 33.75 {
		dir = "NNE"
	} else if nwb >= 33.75 && nwb < 56.25 {
		dir = "NE"
	} else if nwb >= 56.25 && nwb < 78.75 {
		dir = "ENE"
	} else if nwb >= 78.75 && nwb < 101.25 {
		dir = "E"
	} else if nwb >= 101.25 && nwb < 123.75 {
		dir = "ESE"
	} else if nwb >= 123.75 && nwb < 146.25 {
		dir = "SE"
	} else if nwb >= 146.25 && nwb < 168.75 {
		dir = "SSE"
	} else if nwb >= 168.75 && nwb < 191.25 {
		dir = "S"
	} else if nwb >= 191.25 && nwb < 213.75 {
		dir = "SSW"
	} else if nwb >= 213.75 && nwb < 236.25 {
		dir = "SW"
	} else if nwb >= 236.25 && nwb < 258.75 {
		dir = "WSW"
	} else if nwb >= 258.75 && nwb < 281.25 {
		dir = "W"
	} else if nwb >= 281.25 && nwb < 303.75 {
		dir = "WNW"
	} else if nwb >= 303.75 && nwb < 326.25 {
		dir = "NW"
	} else if nwb >= 326.25 && nwb < 348.75 {
		dir = "NNW"
	} else if nwb >= 348.75 {
		dir = "N"
	}

	return fmt.Sprintf("%s (%.1f°)", dir, round(nwb, 0.5))
}

func round(x, unit float64) float64 {
	return math.Round(x/unit) * unit
}

func normalizeBearing(d float64) float64 {
	if d < 0 {
		return normalizeBearing(360 - d)
	} else if d >= 360 {
		return normalizeBearing(d - 360)
	} else {
		return d
	}
}

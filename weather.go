package weather

import (
	"math"
	"os"
	"time"

	"github.com/scheibo/geo"
)

type Client struct {
	provider provider
}

type provider interface {
	current(ll geo.LatLng) (*Conditions, error)
	forecast(ll geo.LatLng) (*Forecast, error)
	history(ll geo.LatLng, t time.Time) (*Conditions, error)
}

type options struct {
	darkSkyKey string
	timezone   *time.Location
}

type Forecast struct {
	Hourly []*Conditions
}

func NewClient(opts ...func(*options)) *Client {
	options := &options{
		darkSkyKey: os.Getenv("DARKSKY_API_KEY"),
		timezone:   time.UTC,
	}

	for _, opt := range opts {
		opt(options)
	}

	return &Client{
		provider: newDarkSkyProvider(options.darkSkyKey, options.timezone),
	}
}

func DarkSky(key string) func(*options) {
	return func(opts *options) {
		if key != "" {
			opts.darkSkyKey = key
		}
	}
}

func TimeZone(loc *time.Location) func(*options) {
	return func(opts *options) {
		if loc != nil {
			opts.timezone = loc
		}
	}
}

func (c *Client) Current(ll geo.LatLng) (*Conditions, error) {
	return c.provider.current(ll)
}

func (c *Client) Now(ll geo.LatLng) (*Conditions, error) {
	return c.Current(ll)
}

func (c *Client) Forecast(ll geo.LatLng) (*Forecast, error) {
	return c.provider.forecast(ll)
}

func (c *Client) History(ll geo.LatLng, t time.Time) (*Conditions, error) {
	return c.provider.history(ll, t)
}

func (c *Client) At(ll geo.LatLng, t time.Time) (*Conditions, error) {
	return c.History(ll, t)
}

func Average(cs []*Conditions) *Conditions {
	n := len(cs)
	if n == 0 {
		return nil
	}

	t0 := time.Time{}

	avg := *cs[0]
	avg.Icon = ""
	avg.Time = t0
	avg.PrecipType = ""
	avg.SunriseTime = t0
	avg.SunsetTime = t0

	var nsws, ewws, nswg, ewwg, wb float64

	for i := 1; i < n; i++ {
		c := cs[i]
		avg.Temperature += c.Temperature
		avg.Humidity += c.Humidity
		avg.ApparentTemperature += c.ApparentTemperature
		avg.PrecipProbability += c.PrecipProbability
		avg.PrecipIntensity += c.PrecipIntensity
		avg.AirPressure += c.AirPressure
		avg.AirDensity += c.AirDensity
		avg.CloudCover += c.CloudCover
		avg.UVIndex += c.UVIndex

		wb = c.WindBearing * geo.DEGREES_TO_RADIANS
		ewws += c.WindSpeed * math.Sin(wb)
		nsws += c.WindSpeed * math.Cos(wb)
		ewwg += c.WindGust * math.Sin(wb)
		nswg += c.WindGust * math.Cos(wb)
	}

	f := float64(n)
	avg.Temperature /= f
	avg.Humidity /= f
	avg.ApparentTemperature /= f
	avg.PrecipProbability /= f
	avg.PrecipIntensity /= f
	avg.AirPressure /= f
	avg.AirDensity /= f
	avg.CloudCover /= f
	avg.UVIndex /= n

	ewws /= f
	nsws /= f
	ewwg /= f
	nswg /= f

	avg.WindSpeed = math.Sqrt(nsws*nsws + ewws*ewws)
	avg.WindGust = math.Sqrt(nswg*nswg + ewwg*ewwg)
	wb = math.Atan2(ewws, nsws)
	if nsws < 0 {
		wb += math.Pi
	}
	avg.WindBearing = normalizeBearing(wb * geo.RADIANS_TO_DEGREES)

	return &avg
}

func rho(t, p, dp float64) float64 {
	const Rd = 287.0531 // specific gas constant for dry air in J(kg*K)
	const Rv = 461.4964 // specific gas constant for water vapor in J(kg*K)
	const K = 273.15    // the value of Kelvin corresponding to 0 Celsius.

	// Herman Wobus constants
	const c0 = 0.99999683
	const c1 = -0.90826951e-02
	const c2 = 0.78736169e-04
	const c3 = -0.61117958e-06
	const c4 = 0.43884187e-08
	const c5 = -0.29883885e-10
	const c6 = 0.21874425e-12
	const c7 = -0.17892321e-14
	const c8 = 0.11112018e-16
	const c9 = -0.30994571e-19

	x := c0 + dp*(c1+dp*(c2+dp*(c3+dp*(c4+dp*(c5+dp*(c6+dp*(c7+dp*(c8+dp*(c9)))))))))
	pv := 6.1078 / (math.Pow(x, 8))

	return 100 * (((p - pv) / (Rd * (t + K))) +
		(pv / (Rv * (t + K))))
}

package weather

import (
	"math"
	"os"
	"time"

	"github.com/scheibo/geo"
)


type Client struct {
	provider *provider
}

type provider interface {
	current(ll geo.LatLng) (*Conditions, error)
	forecast(ll geo.LatLng) (*Forecast, error)
	history(ll geo.LatLng, t time.Time) (*Conditions, error)
}

type Forecast struct {
	Hourly []*Conditions // TODO(kjs): include timestamps?
	Daily []*Conditions
}

func NewClient(opts ...func(*options)) (*Client, error) {
	options := &options{
		ds: keyWeight{ key: os.Getenv("DARKSKY_API_KEY") },
		wu: keyWeight{ key: os.Getenv("WUNDERGROUND_API_KEY"), },
		noaa: keyWeight{ key: os.Getenv("NOAA_API_KEY"), },
		owm: keyWeight{ key: os.Getenv("OWM_API_KEY"), },
	}

	for _, opt := range opts {
		opt(options)
	}
	options = adjustWeights(options)

	provider, err := newEnsembleProvider(options)
	if err != nil {
		return nil, err
	}

	return &Client{provider: provider}, nil
}

func DarkSky(key string, weight ...float64) func(*options) {
	return func(opts *options) {
		opts.ds.key = key
		if len(weight) > 0 {
			opts.ds.weight = weight[0]
		}
	}
}

func Wunderground(key string, weight ...float64) func(*options) {
	return func(opts *options) {
		opts.wu.key = key
		if len(weight) > 0 {
			opts.wu.weight = weight[0]
		}
	}
}

func NOAA(key string, weight ...float64) func(*options) {
	return func(opts *options) {
		opts.noaa.key = key
		if len(weight) > 0 {
			opts.noaa.weight = weight[0]
		}
	}
}

func OpenWeatherMap(key string, weight ...float64) func(*options) {
	return func(opts *options) {
		opts.owm.key = key
		if len(weight) > 0 {
			opts.owm.weight = weight[0]
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

// weather provides a CLI for querying what temperature it was at a particular time
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/scheibo/geo"
	"github.com/scheibo/weather"
)

type TimeFlag struct {
	Time *time.Time
}

func (t *TimeFlag) String() string {
	return fmt.Sprintf("%s", t.Time)
}

func (t *TimeFlag) Set(v string) error {
	parsed, err := dateparse.ParseLocal(strings.TrimSpace(v))
	if err != nil {
		return err
	}
	t.Time = &parsed
	return nil
}

type LatLngFlag struct {
	LatLng *geo.LatLng
}

func (ll *LatLngFlag) String() string {
	return fmt.Sprintf("%s", ll.LatLng)
}

func (ll *LatLngFlag) Set(v string) error {
	latlng, err := geo.ParseLatLng(v)
	if err != nil {
		return err
	}

	ll.LatLng = &latlng
	return nil
}

func main() {
	var key, tz string
	var tf TimeFlag
	var llf LatLngFlag
	var t time.Time
	var ll geo.LatLng

	flag.StringVar(&key, "key", "", "DarkySky API Key")
	flag.StringVar(&tz, "tz", "America/Los_Angeles", "timezone to use")
	flag.Var(&llf, "latlng", "latitude and longitude to query weather information for")
	flag.Var(&tf, "time", "time to query weather information for")

	flag.Parse()

	if tf.Time != nil {
		t = *tf.Time
	} else {
		t = time.Now()
	}

	if llf.LatLng != nil {
		ll = *llf.LatLng
	} else {
		exit(fmt.Errorf("latlng required"))
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		exit(err)
	}

	w := weather.NewClient(weather.DarkSky(key), weather.TimeZone(loc))

	c, err := w.History(ll, t)
	if err != nil {
		exit(err)
	}
	fmt.Println(c)
}

func exit(err error) {
	fmt.Fprintf(os.Stderr, "%s\n\n", err)
	flag.PrintDefaults()
	os.Exit(1)
}

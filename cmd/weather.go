// weather provides a CLI for querying what temperature it was at a particular time
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/araddon/dateparse"
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
	LatLng *weather.LatLng
}

func (ll *LatLngFlag) String() string {
	return fmt.Sprintf("%s", ll.LatLng)
}

func (ll *LatLngFlag) Set(v string) error {
	latlng, err := weather.ParseLatLng(v)
	if err != nil {
		return err
	}

	ll.LatLng = &latlng
	return nil
}

func main() {
	var key string
	var tf TimeFlag
	var llf LatLngFlag
	var t time.Time
	var ll weather.LatLng

	flag.StringVar(&key, "key", "", "DarkySky API Key")
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

	w, err := weather.Get(ll, t, key)
	if err != nil {
		exit(err)
	}
	fmt.Println(w)
}

func exit(err error) {
	fmt.Fprintf(os.Stderr, "%s\n\n", err)
	flag.PrintDefaults()
	os.Exit(1)
}

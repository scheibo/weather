package wtwi

import (
	"fmt"
	"strconv"
	"strings"
)

type LatLng struct {
	Lat float64
	Lng float64
}

func (ll LatLng) String() string {
	return fmt.Sprintf("%.6f,%.6f", ll.Lat, ll.Lng)
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

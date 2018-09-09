package weather

import (
	"fmt"
	"math"
)

const msToKmh = 3600.0 / 1000.0

type Conditions struct {
	//Icon string `json:"icon",omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
	Humidity float64 `json:"humidity,omitempty"`
	PrecipProbability float64 `json:"precipProbability,omitempty"`
	PrecipIntensity float64 `json:"precipIntensity,omitempty"`
	AirPressure float64 `json:"airPressure,omitempty"`
	AirDensity  float64 `json:"airDensity,omitempty"`
	WindSpeed   float64 `json:"windSpeed,omitempty"`
	WindGust   float64 `json:"windGust,omitempty"`
	WindBearing float64 `json:"windBearing,omitempty"`
	UVIndex float64 `json:"uvIndex,omitempty"`
}

func (c *Conditions) String() string {
	return fmt.Sprintf(
		"temp: %.1f °C (humidity: %d%%)" +
		"precip: %d%% %.3f mm/h" +
		"wind: %.1f km/h (gust: %.1f km/h) %s" +
		"density: %.3f kg/m³ (pressure: %.2f mbar)" +
		"uvIndex: %d",
		round(c.Temperature, 0.1),
		int(c.Humidity * 100),
		int(c.PrecipProbability * 100),
		round(c.PreciptIntensity, 0.001),
		round(c.WindSpeed*msToKmh, 0.1),
		round(c.WindGust*msToKmh, 0.1),
		bearingString(c.WindBearing),
		round(c.AirDensity, 0.001),
		round(c.AirPressure, 0.01)
		int(c.UvIndex))
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

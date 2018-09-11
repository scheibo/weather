package weather

import (
	"fmt"
	"math"
	"time"
)

const msToKmh = 3600.0 / 1000.0

type Conditions struct {
	Icon              string    `json:"icon",omitempty"`
	Time              time.Time `json:"time",omitempty"`
	Temperature       float64   `json:"temperature,omitempty"`
	Humidity          float64   `json:"humidity,omitempty"`
	PrecipProbability float64   `json:"precipProbability,omitempty"`
	PrecipIntensity   float64   `json:"precipIntensity,omitempty"`
	AirPressure       float64   `json:"airPressure,omitempty"`
	AirDensity        float64   `json:"airDensity,omitempty"`
	CloudCover        float64   `json:"cloudCover,omitempty"`
	UVIndex           int       `json:"uvIndex,omitempty"`
	WindSpeed         float64   `json:"windSpeed,omitempty"`
	WindGust          float64   `json:"windGust,omitempty"`
	WindBearing       float64   `json:"windBearing,omitempty"`
}

func (c *Conditions) String() string {
	precip := ""
	if c.PrecipProbability > 0 && c.PrecipIntensity > 0 {
		precip = fmt.Sprintf("precip: %d%% %.3f mm/h\n",
			int(c.PrecipProbability*100),
			round(c.PrecipIntensity, 0.001))
	}

	return fmt.Sprintf(
		"%v\n"+
			"temp: %.1f °C (humidity: %d%%)\n"+
			"%s"+
			"wind: %.1f km/h (gust: %.1f km/h) %s\n"+
			"density: %.3f kg/m³ (pressure: %.2f mbar)\n"+
			"%s uv: %d cloud: %d%%",
		c.Time,
		round(c.Temperature, 0.1),
		int(c.Humidity*100),
		precip,
		round(c.WindSpeed*msToKmh, 0.1),
		round(c.WindGust*msToKmh, 0.1),
		bearingString(c.WindBearing),
		round(c.AirDensity, 0.001),
		round(c.AirPressure, 0.01),
		c.Icon,
		c.UVIndex,
		int(c.CloudCover*100))
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

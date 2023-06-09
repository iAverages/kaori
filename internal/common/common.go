package common

type ErrorResponse struct {
	Message string `json:"message"`
}

type Song struct {
	Artist   Artist `json:"artist,omitempty"`
	Name     string `json:"name,omitempty"`
	Duration int    `json:"duration,omitempty"`
	Url      string `json:"url,omitempty"`
}

type Artist struct {
	Name string `json:"name,omitempty"`
	Url  string `json:"url,omitempty"`
}

type PlayingNow struct {
	Song        *Song     `json:"song,omitempty"`
	IsPlaying   bool      `json:"is_playing"`
	Progress    int       `json:"progress,omitempty"`
	PlaylistUrl string    `json:"playlist_url,omitempty"`
	Icon        string    `json:"icon,omitempty"`
	Levels      []float64 `json:"levels,omitempty"`
}

type AudioAnalysis struct {
	Start    float64 `json:"start,omitempty"`
	Duration float64 `json:"duration,omitempty"`
	Loudness float64 `json:"loudness,omitempty"`
}

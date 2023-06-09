package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"kaori/internal/common"
	"kaori/internal/config"
	"kaori/internal/redis"
	"math"
	"net/http"

	"github.com/uptrace/bunrouter"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"go.step.sm/crypto/randutil"
	"go.uber.org/zap"
)

var (
	ch    = make(chan *spotify.Client)
	state string
)
var client *spotify.Client
var auth *spotifyauth.Authenticator

type Service struct {
	logger *zap.SugaredLogger
	config *config.Config
	rdb    redis.RedisService
}

func NewSpotifyService(logger *zap.SugaredLogger, config *config.Config, rdb redis.RedisService) *Service {
	return &Service{logger, config, rdb}
}

func (service *Service) Init() *spotify.PrivateUser {
	cfg := service.config
	logger := service.logger
	rdb := service.rdb

	auth = spotifyauth.New(
		spotifyauth.WithRedirectURL(cfg.Hostname+"/callback"),
		spotifyauth.WithScopes(spotifyauth.ScopeUserReadCurrentlyPlaying, spotifyauth.ScopeUserReadPlaybackState, spotifyauth.ScopeUserModifyPlaybackState),
		spotifyauth.WithClientID(cfg.SpofityId),
		spotifyauth.WithClientSecret(cfg.SpofitySecret),
	)

	token, err := rdb.GetLastToken()
	if err == nil {
		client = spotify.New(auth.Client(context.Background(), token))
		// Get current user to check if token is valid still
		user, err := client.CurrentUser(context.Background())
		if err != nil {
			logger.Error(err)
			return nil
		}
		return user
	} else {
		rad, err := randutil.Alphanumeric(16)
		if err != nil {
			logger.Error(err)
		}
		state = rad
	}
	return nil
}

func (service *Service) DisplayAuthURL() {
	url := auth.AuthURL(state)
	service.logger.Info("Please log in to Spotify by visiting the following page in your browser:", url)

	// wait for auth to complete
	client = <-ch

	// use the client to make calls that require authorization
	user, err := client.CurrentUser(context.Background())
	if err != nil {
		service.logger.Fatal(err)
	}
	service.logger.Info("user id:", user.ID)
}

func findSegment(segments []common.AudioAnalysis, i float64) *common.AudioAnalysis {
	for _, segment := range segments {
		if i <= segment.Start+segment.Duration {
			return &segment
		}
	}
	return nil
}

func max(arr []common.AudioAnalysis) float64 {
	var max float64
	for _, v := range arr {
		if v.Loudness > max {
			max = v.Loudness
		}
	}
	return max
}

func formatAnalysis(analysis *spotify.AudioAnalysis) []float64 {

	var segments []common.AudioAnalysis
	duration := analysis.Track.Duration

	for _, segment := range analysis.Segments {
		segments = append(segments, common.AudioAnalysis{
			Start:    segment.Start / float64(duration),
			Duration: segment.Duration / float64(duration),
			Loudness: 1 - (math.Min(math.Max(segment.LoudnessStart, -35), 0) / -35),
		})
	}

	max := max(segments)
	var levels = []float64{}

	for i := 0.000; i < 1; i += 0.001 {
		segment := findSegment(segments, i)
		if segment == nil {
			levels = append(levels, 0)
			continue
		}
		loudness := math.Round((segment.Loudness/max)*100) / 100
		levels = append(levels, loudness)
	}

	return levels
}

func (service *Service) GetCurrentSong() common.PlayingNow {
	if client == nil {
		service.logger.Error("Client is nil")
		return common.PlayingNow{
			IsPlaying: false,
			Song:      nil,
		}
	}

	cache, err := service.rdb.GetSongCache()

	if err == nil {
		return cache
	}

	playerState, err := client.PlayerState(context.Background())
	if playerState == nil {
		service.logger.Error("PlayerState is nil")
		return common.PlayingNow{
			IsPlaying: false,
			Song:      nil,
		}
	}

	if err != nil {
		service.logger.Error(err)
		return common.PlayingNow{
			IsPlaying: false,
			Song:      nil,
		}
	}

	if playerState.Playing && playerState.Item != nil {
		analysis, err := client.GetAudioAnalysis(context.Background(), playerState.Item.ID)

		if err != nil {
			service.logger.Error(err)
		}

		playing := common.PlayingNow{
			Song: &common.Song{
				Artist: common.Artist{
					Name: playerState.Item.Artists[0].Name,
					Url:  playerState.Item.Artists[0].ExternalURLs["spotify"],
				},
				Name:     playerState.Item.Name,
				Duration: playerState.Item.Duration,
				Url:      playerState.Item.ExternalURLs["spotify"],
			},
			PlaylistUrl: playerState.PlaybackContext.ExternalURLs["spotify"],
			IsPlaying:   playerState.Playing,
			Progress:    playerState.Progress,
			Icon:        playerState.Item.Album.Images[0].URL,
			Levels:      formatAnalysis(analysis),
		}
		go func() {
			err := service.rdb.SaveSongCache(playing)
			if err != nil {
				service.logger.Error(err)
			}
		}()

		return playing

	}

	return common.PlayingNow{
		IsPlaying: false,
		Song:      nil,
	}
}

func (service *Service) Callback(w http.ResponseWriter, req bunrouter.Request) error {
	tok, err := auth.Token(req.Request.Context(), state, req.Request)
	if err != nil {
		http.Error(w, "404 page not found", http.StatusNotFound)
		return nil
	}

	if st := req.Request.FormValue("state"); st != state {
		http.NotFound(w, req.Request)
		service.logger.Fatalf("State mismatch: %s != %s\n", st, state)
	}

	client := spotify.New(auth.Client(req.Request.Context(), tok))
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "Authorization Complete")
	ch <- client

	json, err := json.Marshal(tok)
	if err != nil {
		service.logger.Fatal(err)
	}

	err = service.rdb.SaveToken(json)
	if err != nil {
		service.logger.Fatal(err)
	}

	return nil
}

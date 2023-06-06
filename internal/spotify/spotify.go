package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"kaori/internal/config"
	"kaori/internal/redis"
	"log"
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

func Init(cfg *config.Config, logger *zap.SugaredLogger) *spotify.PrivateUser {

	redis.Init(cfg)

	auth = spotifyauth.New(
		spotifyauth.WithRedirectURL(cfg.Hostname+"/callback"),
		spotifyauth.WithScopes(spotifyauth.ScopeUserReadCurrentlyPlaying, spotifyauth.ScopeUserReadPlaybackState, spotifyauth.ScopeUserModifyPlaybackState),
		spotifyauth.WithClientID(cfg.SpofityId),
		spotifyauth.WithClientSecret(cfg.SpofitySecret),
	)

	token, err := redis.GetLastToken()
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

func DisplayAuthURL(cfg *config.Config, logger *zap.SugaredLogger) {
	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	// wait for auth to complete
	client = <-ch

	// use the client to make calls that require authorization
	user, err := client.CurrentUser(context.Background())
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info("user id:", user.ID)
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
	Song        *Song  `json:"song,omitempty"`
	IsPlaying   bool   `json:"is_playing"`
	Progress    int    `json:"progress,omitempty"`
	PlaylistUrl string `json:"playlist_url,omitempty"`
}

func GetCurrentSong(logger *zap.SugaredLogger) PlayingNow {
	if client == nil {
		logger.Error("Client is nil")
		return PlayingNow{
			IsPlaying: false,
			Song:      nil,
		}
	}

	playerState, err := client.PlayerState(context.Background())
	if playerState == nil {
		logger.Error("PlayerState is nil")
		return PlayingNow{
			IsPlaying: false,
			Song:      nil,
		}
	}

	if err != nil {
		logger.Error(err)
		return PlayingNow{
			IsPlaying: false,
			Song:      nil,
		}
	}
	if playerState.Playing && playerState.Item != nil {
		return PlayingNow{
			Song: &Song{
				Artist: Artist{
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
		}
	}

	return PlayingNow{
		IsPlaying: false,
		Song:      nil,
	}
}

func Callback(w http.ResponseWriter, req bunrouter.Request) error {
	tok, err := auth.Token(req.Request.Context(), state, req.Request)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}

	if st := req.Request.FormValue("state"); st != state {
		http.NotFound(w, req.Request)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}

	client := spotify.New(auth.Client(req.Request.Context(), tok))
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "Authorization Complete")
	ch <- client

	json, err := json.Marshal(tok)
	if err != nil {
		log.Fatal(err)
	}

	err = redis.SaveToken(json)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

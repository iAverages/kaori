package app

import (
	"kaori/internal/config"
	"kaori/internal/spotify"
	"net/http"

	"github.com/uptrace/bunrouter"
	"go.uber.org/zap"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

var isReady = false

func serverIsReady(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		if !isReady {
			return bunrouter.JSON(w, ErrorResponse{Message: "Kaori not ready yet"})
		}
		return next(w, req)
	}
}

func corsMiddleware(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		h := w.Header()

		h.Set("Access-Control-Allow-Origin", "*")

		// CORS preflight.
		if req.Method == http.MethodOptions {
			h.Set("Access-Control-Allow-Methods", "GET")
			h.Set("Access-Control-Allow-Headers", "content-type")
			h.Set("Access-Control-Max-Age", "86400")
			return nil
		}

		return next(w, req)
	}
}

func Start(cfg *config.Config, logger *zap.SugaredLogger) {
	logger.Info("Starting server...")
	spotifyService := spotify.NewService(logger, cfg)

	router := bunrouter.New(bunrouter.WithMiddleware(corsMiddleware))

	router.GET("/callback", spotifyService.Callback)

	group := router.NewGroup("/api").Use(serverIsReady)

	group.GET("/ping", func(w http.ResponseWriter, req bunrouter.Request) error {
		w.Write([]byte("pong"))
		return nil
	})

	group.GET("/player", func(w http.ResponseWriter, req bunrouter.Request) error {
		return bunrouter.JSON(w, spotifyService.GetCurrentSong())
	})

	go func() {
		user := spotifyService.Init()
		if user == nil {
			spotifyService.DisplayAuthURL()
		} else {
			logger.Info("You are logged in as: ", user.DisplayName, "(", user.ID, ")")
		}
		isReady = true
	}()

	logger.Info("Server started on port ", cfg.Port)
	err := http.ListenAndServe("0.0.0.0:"+cfg.Port, router)
	if err != nil {
		logger.Fatal(err)
	}
}

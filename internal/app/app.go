package app

import (
	"kaori/internal/config"
	"kaori/internal/spotify"
	"log"
	"net/http"

	"github.com/uptrace/bunrouter"
	"go.uber.org/zap"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func Start(cfg *config.Config, logger *zap.SugaredLogger) {
	logger.Info("Starting server...")

	router := bunrouter.New()

	router.GET("/ping", func(w http.ResponseWriter, req bunrouter.Request) error {
		w.Write([]byte("pong"))
		return nil
	})

	router.GET("/player", func(w http.ResponseWriter, req bunrouter.Request) error {
		return bunrouter.JSON(w, spotify.GetCurrentSong(logger))
	})

	router.GET("/callback", spotify.Callback)
	logger.Info("Server started on port ", cfg.Port)
	go func() {
		user := spotify.Init(cfg, logger)
		if user == nil {
			spotify.DisplayAuthURL(cfg, logger)
		} else {
			logger.Info("You are logged in as:", user.DisplayName, "(", user.ID, ")")
		}
	}()

	err := http.ListenAndServe("0.0.0.0:"+cfg.Port, router)
	if err != nil {
		log.Fatal(err)
	}
}

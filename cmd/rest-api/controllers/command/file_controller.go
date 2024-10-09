package cmd_controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	replay_in "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/ports/in"
)

type FileController struct {
	container container.Container
}

func NewFileController(container container.Container) *FileController {
	return &FileController{container: container}
}

func (ctlr *FileController) UploadHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "localhost:4991") // todo: PARAMETRIZAR
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		// r.Body = http.MaxBytesReader(w, r.Body, 32<<20)
		r.ParseMultipartForm(32 << 50)

		reqContext := context.WithValue(r.Context(), common.GameIDParamKey, r.FormValue("game_id"))

		file, _, err := r.FormFile("file")
		if err != nil {
			slog.ErrorContext(reqContext, "Failed to get file", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer file.Close()

		var uploadAndProcessReplayFileCommand replay_in.UploadAndProcessReplayFileCommand
		err = ctlr.container.Resolve(&uploadAndProcessReplayFileCommand)
		if err != nil {
			slog.ErrorContext(reqContext, "Failed to resolve uploadAndProcessReplayFileCommand", "err", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		match, err := uploadAndProcessReplayFileCommand.Exec(reqContext, file)
		if err != nil {
			slog.ErrorContext(reqContext, "Failed to upload and process file", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		match.Events = nil

		err = json.NewEncoder(w).Encode(match)
		if err != nil {
			slog.ErrorContext(reqContext, "Failed to encode response", "err", err, "match", match)
			w.WriteHeader(http.StatusBadGateway)
		}

		w.Header().Set("Location", r.URL.Path+"/"+match.ID.String())
		w.WriteHeader(http.StatusCreated)
	}
}

// func (ctlr *FileController) ReplayMetadataFilterHandler(apiContext context.Context) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Set("Access-Control-Allow-Origin", "localhost:3000")
// 		w.Header().Set("Access-Control-Allow-Methods", "GET")
// 		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

// 		reqContext := context.WithValue(r.Context(), common.GameIDParamKey, r.FormValue("game_id"))

// 		var replayFileMetadataReader replay_in.ReplayFileMetadataReader
// 		err := ctlr.container.Resolve(&replayFileMetadataReader)
// 		if err != nil {
// 			slog.ErrorContext(reqContext, "Failed to resolve replayFileMetadataReader", "err", err)
// 			w.WriteHeader(http.StatusServiceUnavailable)
// 			return
// 		}

// 		var params []common.SearchAggregation

// 		// for key, values := range r.URL.Query() {
// 		// 	params = append(params, common.SearchAggregation{
// 		// 		Key:    key,
// 		// 		Values: values,
// 		// 	})
// 		// }

// 		// replayFiles, err := replayFileMetadataReader.Filter(reqContext, r.URL.Query())
// 		// if err != nil {
// 		// 	slog.ErrorContext(reqContext, "Failed to get replay files", "err", err)
// 		// 	w.WriteHeader(http.StatusInternalServerError)
// 		// 	return
// 		// }

// 		// err = json.NewEncoder(w).Encode(replayFiles)
// 		// if err != nil {
// 		// 	slog.ErrorContext(reqContext, "Failed to encode response", err, "replayFiles", replayFiles)
// 		// 	w.WriteHeader(http.StatusBadGateway)
// 		// }

// 		// w.Header().Set("Location", r.URL.Path)
// 		// w.WriteHeader(http.StatusOK)
// 	}
// }

package cmd_controllers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/replay-api/replay-api/cmd/rest-api/controllers"
	common "github.com/replay-api/replay-api/pkg/domain"
	replay_in "github.com/replay-api/replay-api/pkg/domain/replay/ports/in"
)

type FileController struct {
	container container.Container
	helper    *controllers.ControllerHelper
}

func NewFileController(container container.Container) *FileController {
	return &FileController{
		container: container,
		helper:    controllers.NewControllerHelper(),
	}
}

func (ctlr *FileController) UploadHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // todo: PARAMETRIZAR
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		// r.Body = http.MaxBytesReader(w, r.Body, 32<<57)
		r.ParseMultipartForm(32 << 50)

		reqContext := context.WithValue(r.Context(), common.GameIDParamKey, r.FormValue("game_id"))

		slog.InfoContext(reqContext, "Receiving file", string(common.GameIDParamKey), r.FormValue("game_id"))

		file, _, err := r.FormFile("file")
		if err != nil {
			slog.ErrorContext(reqContext, "Failed to get file", "err", err)
			ctlr.helper.HandleError(w, r, common.NewAPIError(http.StatusBadRequest, "BAD_REQUEST", "failed to get file from form"), "form file error")
			return
		}
		defer file.Close()

		var uploadAndProcessReplayFileCommand replay_in.UploadAndProcessReplayFileCommand
		err = ctlr.container.Resolve(&uploadAndProcessReplayFileCommand)
		if err != nil {
			slog.ErrorContext(reqContext, "Failed to resolve uploadAndProcessReplayFileCommand", "err", err)
			ctlr.helper.HandleError(w, r, common.NewAPIError(http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "failed to resolve command handler"), "dependency resolution failed")
			return
		}

		match, err := uploadAndProcessReplayFileCommand.Exec(reqContext, file)
		if err != nil {
			// Handle specific error cases
			if err.Error() == "Unauthorized" {
				ctlr.helper.HandleError(w, r, common.NewAPIError(http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized access"), "unauthorized access")
			} else {
				ctlr.helper.HandleError(w, r, err, "Failed to upload and process file")
			}
			return
		}

		// Remove events from the response (business requirement)
		match.Events = nil

		// Set Location header and write successful response
		w.Header().Set("Location", r.URL.Path+"/"+match.ID.String())
		ctlr.helper.WriteCreated(w, r, match)
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

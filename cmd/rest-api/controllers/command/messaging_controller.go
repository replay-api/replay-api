package cmd_controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	messaging_entities "github.com/replay-api/replay-api/pkg/domain/messaging/entities"
	messaging_in "github.com/replay-api/replay-api/pkg/domain/messaging/ports/in"
)

type MessagingCommandController struct {
	commentCommand       messaging_in.CommentCommand
	directMessageCommand messaging_in.DirectMessageCommand
	teamMessageCommand   messaging_in.TeamMessageCommand
}

func NewMessagingCommandController(c container.Container) *MessagingCommandController {
	var commentCommand messaging_in.CommentCommand
	var directMessageCommand messaging_in.DirectMessageCommand
	var teamMessageCommand messaging_in.TeamMessageCommand

	if err := c.Resolve(&commentCommand); err != nil {
		slog.Warn("CommentCommand not available", "error", err)
	}
	if err := c.Resolve(&directMessageCommand); err != nil {
		slog.Warn("DirectMessageCommand not available", "error", err)
	}
	if err := c.Resolve(&teamMessageCommand); err != nil {
		slog.Warn("TeamMessageCommand not available", "error", err)
	}

	return &MessagingCommandController{
		commentCommand:       commentCommand,
		directMessageCommand: directMessageCommand,
		teamMessageCommand:   teamMessageCommand,
	}
}

// ==================== Comment Handlers ====================

// CreateCommentHandler handles POST /matches/{match_id}/comments
func (ctrl *MessagingCommandController) CreateCommentHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		matchIDStr := vars["match_id"]

		matchID, err := uuid.Parse(matchIDStr)
		if err != nil {
			http.Error(w, `{"success":false,"error":"invalid match_id"}`, http.StatusBadRequest)
			return
		}

		var req struct {
			Content  string                       `json:"content"`
			Mentions []messaging_entities.Mention `json:"mentions,omitempty"`
			ParentID *string                      `json:"parent_id,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"success":false,"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}

		cmd := messaging_in.CreateCommentCommand{
			MatchID:  matchID,
			Content:  req.Content,
			Mentions: req.Mentions,
		}

		if req.ParentID != nil {
			parentID, err := uuid.Parse(*req.ParentID)
			if err != nil {
				http.Error(w, `{"success":false,"error":"invalid parent_id"}`, http.StatusBadRequest)
				return
			}
			cmd.ParentID = &parentID
		}

		if err := cmd.Validate(); err != nil {
			http.Error(w, `{"success":false,"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		comment, err := ctrl.commentCommand.CreateComment(r.Context(), cmd)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to create comment", "error", err)
			http.Error(w, `{"success":false,"error":"failed to create comment"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(comment)
	}
}

// EditCommentHandler handles PUT /matches/{match_id}/comments/{comment_id}
func (ctrl *MessagingCommandController) EditCommentHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		commentIDStr := vars["comment_id"]

		commentID, err := uuid.Parse(commentIDStr)
		if err != nil {
			http.Error(w, `{"success":false,"error":"invalid comment_id"}`, http.StatusBadRequest)
			return
		}

		var req struct {
			Content  string                       `json:"content"`
			Mentions []messaging_entities.Mention `json:"mentions,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"success":false,"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}

		cmd := messaging_in.EditCommentCommand{
			CommentID: commentID,
			Content:   req.Content,
			Mentions:  req.Mentions,
		}

		if err := cmd.Validate(); err != nil {
			http.Error(w, `{"success":false,"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		comment, err := ctrl.commentCommand.EditComment(r.Context(), cmd)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to edit comment", "error", err)
			http.Error(w, `{"success":false,"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(comment)
	}
}

// DeleteCommentHandler handles DELETE /matches/{match_id}/comments/{comment_id}
func (ctrl *MessagingCommandController) DeleteCommentHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		commentIDStr := vars["comment_id"]

		commentID, err := uuid.Parse(commentIDStr)
		if err != nil {
			http.Error(w, `{"success":false,"error":"invalid comment_id"}`, http.StatusBadRequest)
			return
		}

		cmd := messaging_in.DeleteCommentCommand{
			CommentID: commentID,
		}

		if err := ctrl.commentCommand.DeleteComment(r.Context(), cmd); err != nil {
			slog.ErrorContext(r.Context(), "failed to delete comment", "error", err)
			http.Error(w, `{"success":false,"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// ReactToCommentHandler handles POST /matches/{match_id}/comments/{comment_id}/reactions
func (ctrl *MessagingCommandController) ReactToCommentHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		commentIDStr := vars["comment_id"]

		commentID, err := uuid.Parse(commentIDStr)
		if err != nil {
			http.Error(w, `{"success":false,"error":"invalid comment_id"}`, http.StatusBadRequest)
			return
		}

		var req struct {
			Emoji  string `json:"emoji"`
			Remove bool   `json:"remove,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"success":false,"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}

		cmd := messaging_in.ReactToCommentCommand{
			CommentID: commentID,
			Emoji:     req.Emoji,
			Remove:    req.Remove,
		}

		if err := cmd.Validate(); err != nil {
			http.Error(w, `{"success":false,"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		if err := ctrl.commentCommand.ReactToComment(r.Context(), cmd); err != nil {
			slog.ErrorContext(r.Context(), "failed to react to comment", "error", err)
			http.Error(w, `{"success":false,"error":"failed to react"}`, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

// ==================== Direct Message Handlers ====================

// SendDirectMessageHandler handles POST /messages/direct
func (ctrl *MessagingCommandController) SendDirectMessageHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			RecipientID string                       `json:"recipient_id"`
			Content     string                       `json:"content"`
			Mentions    []messaging_entities.Mention `json:"mentions,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"success":false,"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}

		recipientID, err := uuid.Parse(req.RecipientID)
		if err != nil {
			http.Error(w, `{"success":false,"error":"invalid recipient_id"}`, http.StatusBadRequest)
			return
		}

		cmd := messaging_in.SendDirectMessageCommand{
			RecipientID: recipientID,
			Content:     req.Content,
			Mentions:    req.Mentions,
		}

		if err := cmd.Validate(); err != nil {
			http.Error(w, `{"success":false,"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		dm, err := ctrl.directMessageCommand.SendDirectMessage(r.Context(), cmd)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to send direct message", "error", err)
			http.Error(w, `{"success":false,"error":"failed to send message"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(dm)
	}
}

// MarkConversationReadHandler handles PUT /messages/conversations/{user_id}/read
func (ctrl *MessagingCommandController) MarkConversationReadHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		otherUserIDStr := vars["user_id"]

		otherUserID, err := uuid.Parse(otherUserIDStr)
		if err != nil {
			http.Error(w, `{"success":false,"error":"invalid user_id"}`, http.StatusBadRequest)
			return
		}

		cmd := messaging_in.MarkConversationReadCommand{
			OtherUserID: otherUserID,
		}

		if err := ctrl.directMessageCommand.MarkConversationRead(r.Context(), cmd); err != nil {
			slog.ErrorContext(r.Context(), "failed to mark conversation read", "error", err)
			http.Error(w, `{"success":false,"error":"failed to mark as read"}`, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "marked_read"})
	}
}

// DeleteDirectMessageHandler handles DELETE /messages/{message_id}
func (ctrl *MessagingCommandController) DeleteDirectMessageHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		messageIDStr := vars["message_id"]

		messageID, err := uuid.Parse(messageIDStr)
		if err != nil {
			http.Error(w, `{"success":false,"error":"invalid message_id"}`, http.StatusBadRequest)
			return
		}

		if err := ctrl.directMessageCommand.DeleteMessage(r.Context(), messageID); err != nil {
			slog.ErrorContext(r.Context(), "failed to delete message", "error", err)
			http.Error(w, `{"success":false,"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// ==================== Team Message Handlers ====================

// SendTeamMessageHandler handles POST /teams/{team_id}/messages
func (ctrl *MessagingCommandController) SendTeamMessageHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		teamIDStr := vars["team_id"]

		teamID, err := uuid.Parse(teamIDStr)
		if err != nil {
			http.Error(w, `{"success":false,"error":"invalid team_id"}`, http.StatusBadRequest)
			return
		}

		var req struct {
			Channel  string                       `json:"channel"`
			Content  string                       `json:"content"`
			Mentions []messaging_entities.Mention `json:"mentions,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"success":false,"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}

		cmd := messaging_in.SendTeamMessageCommand{
			TeamID:   teamID,
			Channel:  messaging_entities.ChannelType(req.Channel),
			Content:  req.Content,
			Mentions: req.Mentions,
		}

		if err := cmd.Validate(); err != nil {
			http.Error(w, `{"success":false,"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		msg, err := ctrl.teamMessageCommand.SendTeamMessage(r.Context(), cmd)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to send team message", "error", err)
			http.Error(w, `{"success":false,"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(msg)
	}
}

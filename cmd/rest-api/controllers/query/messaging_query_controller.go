package query_controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	messaging_entities "github.com/replay-api/replay-api/pkg/domain/messaging/entities"
	messaging_in "github.com/replay-api/replay-api/pkg/domain/messaging/ports/in"
)

type MessagingQueryController struct {
	commentQuery       messaging_in.CommentQuery
	directMessageQuery messaging_in.DirectMessageQuery
	teamMessageQuery   messaging_in.TeamMessageQuery
}

func NewMessagingQueryController(c container.Container) *MessagingQueryController {
	var commentQuery messaging_in.CommentQuery
	var directMessageQuery messaging_in.DirectMessageQuery
	var teamMessageQuery messaging_in.TeamMessageQuery

	if err := c.Resolve(&commentQuery); err != nil {
		slog.Warn("CommentQuery not available", "error", err)
	}
	if err := c.Resolve(&directMessageQuery); err != nil {
		slog.Warn("DirectMessageQuery not available", "error", err)
	}
	if err := c.Resolve(&teamMessageQuery); err != nil {
		slog.Warn("TeamMessageQuery not available", "error", err)
	}

	return &MessagingQueryController{
		commentQuery:       commentQuery,
		directMessageQuery: directMessageQuery,
		teamMessageQuery:   teamMessageQuery,
	}
}

// ==================== Comment Query Handlers ====================

// ListMatchCommentsHandler handles GET /matches/{match_id}/comments
func (ctrl *MessagingQueryController) ListMatchCommentsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchIDStr := vars["match_id"]

	matchID, err := uuid.Parse(matchIDStr)
	if err != nil {
		http.Error(w, `{"success":false,"error":"invalid match_id"}`, http.StatusBadRequest)
		return
	}

	limit, offset := parseMessagingPagination(r)
	sort := r.URL.Query().Get("sort")

	query := messaging_in.ListMatchCommentsQuery{
		MatchID: matchID,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}

	result, err := ctrl.commentQuery.ListMatchComments(r.Context(), query)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to list comments", "error", err)
		http.Error(w, `{"success":false,"error":"failed to list comments"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetCommentHandler handles GET /matches/{match_id}/comments/{comment_id}
func (ctrl *MessagingQueryController) GetCommentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentIDStr := vars["comment_id"]

	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		http.Error(w, `{"success":false,"error":"invalid comment_id"}`, http.StatusBadRequest)
		return
	}

	comment, err := ctrl.commentQuery.GetComment(r.Context(), commentID)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to get comment", "error", err)
		http.Error(w, `{"success":false,"error":"comment not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comment)
}

// GetCommentRepliesHandler handles GET /matches/{match_id}/comments/{comment_id}/replies
func (ctrl *MessagingQueryController) GetCommentRepliesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentIDStr := vars["comment_id"]

	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		http.Error(w, `{"success":false,"error":"invalid comment_id"}`, http.StatusBadRequest)
		return
	}

	limit, offset := parseMessagingPagination(r)

	result, err := ctrl.commentQuery.GetCommentReplies(r.Context(), commentID, limit, offset)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to get replies", "error", err)
		http.Error(w, `{"success":false,"error":"failed to get replies"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ==================== Direct Message Query Handlers ====================

// ListConversationsHandler handles GET /messages/conversations
func (ctrl *MessagingQueryController) ListConversationsHandler(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseMessagingPagination(r)

	query := messaging_in.ListConversationsQuery{
		Limit:  limit,
		Offset: offset,
	}

	result, err := ctrl.directMessageQuery.ListConversations(r.Context(), query)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to list conversations", "error", err)
		http.Error(w, `{"success":false,"error":"failed to list conversations"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetConversationHandler handles GET /messages/conversations/{user_id}
func (ctrl *MessagingQueryController) GetConversationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	otherUserIDStr := vars["user_id"]

	otherUserID, err := uuid.Parse(otherUserIDStr)
	if err != nil {
		http.Error(w, `{"success":false,"error":"invalid user_id"}`, http.StatusBadRequest)
		return
	}

	limit, offset := parseMessagingPagination(r)

	query := messaging_in.GetConversationQuery{
		OtherUserID: otherUserID,
		Limit:       limit,
		Offset:      offset,
	}

	result, err := ctrl.directMessageQuery.GetConversation(r.Context(), query)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to get conversation", "error", err)
		http.Error(w, `{"success":false,"error":"failed to get conversation"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ==================== Team Message Query Handlers ====================

// ListTeamMessagesHandler handles GET /teams/{team_id}/messages
func (ctrl *MessagingQueryController) ListTeamMessagesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamIDStr := vars["team_id"]

	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		http.Error(w, `{"success":false,"error":"invalid team_id"}`, http.StatusBadRequest)
		return
	}

	channel := r.URL.Query().Get("channel")
	if channel == "" {
		channel = "general"
	}

	limit, offset := parseMessagingPagination(r)

	query := messaging_in.ListTeamMessagesQuery{
		TeamID:  teamID,
		Channel: messaging_entities.ChannelType(channel),
		Limit:   limit,
		Offset:  offset,
	}

	result, err := ctrl.teamMessageQuery.ListTeamMessages(r.Context(), query)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to list team messages", "error", err)
		http.Error(w, `{"success":false,"error":"failed to list messages"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ListTeamChannelsHandler handles GET /teams/{team_id}/channels
func (ctrl *MessagingQueryController) ListTeamChannelsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamIDStr := vars["team_id"]

	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		http.Error(w, `{"success":false,"error":"invalid team_id"}`, http.StatusBadRequest)
		return
	}

	channels, err := ctrl.teamMessageQuery.ListTeamChannels(r.Context(), teamID)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to list team channels", "error", err)
		http.Error(w, `{"success":false,"error":"failed to list channels"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"team_id":  teamID.String(),
		"channels": channels,
	})
}

// ==================== Helpers ====================

func parseMessagingPagination(r *http.Request) (int, int) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	return limit, offset
}

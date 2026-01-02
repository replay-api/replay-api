package query_controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	iam_in "github.com/replay-api/replay-api/pkg/domain/iam/ports/in"
)

type GroupController struct {
	MembershipReader iam_in.MembershipReader
}

func NewGroupController(c *container.Container) *GroupController {
	var membershipReader iam_in.MembershipReader
	err := c.Resolve(&membershipReader)

	if err != nil {
		slog.Warn("MembershipReader not available - group queries will be disabled", "error", err)
	}

	return &GroupController{
		MembershipReader: membershipReader,
	}
}

func (c *GroupController) HandleListMemberGroups(w http.ResponseWriter, r *http.Request) {
	s := InitializeSearch(r)

	s = GetQueryParams(r, s)

	groupsAndMemberships, err := c.MembershipReader.ListMemberGroups(r.Context(), s)

	if err != nil {
		slog.Error("Error listing member groups", "error", err, "search", s)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(groupsAndMemberships) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(groupsAndMemberships)
}

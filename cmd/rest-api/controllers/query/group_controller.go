package query_controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	iam_in "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/ports/in"
)

type GroupController struct {
	MembershipReader iam_in.MembershipReader
}

func NewGroupController(c *container.Container) *GroupController {
	var membershipReader iam_in.MembershipReader
	err := c.Resolve(&membershipReader)

	if err != nil {
		slog.Error("Cannot resolve iam_entities.MembershipReader for NewGroupController", "err", err)
		panic(err)
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
	json.NewEncoder(w).Encode(groupsAndMemberships)
}

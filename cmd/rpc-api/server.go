package main

import (
	"context"
	"log/slog"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/replay-api/replay-api/pkg/app/proto/billing/generated/billing"
	"github.com/replay-api/replay-api/pkg/app/proto/iam/generated/iam"
	"github.com/replay-api/replay-api/pkg/app/proto/squad/generated/squad"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type rpcServer struct {
	Container container.Container
	iam.UnimplementedRIDServiceServer
	squad.UnimplementedPlayerProfileServiceServer
	billing.UnimplementedSubscriptionServiceServer

	userTokenCache          map[uuid.UUID]map[uuid.UUID]iam_entities.RIDToken
	tokenUserCache          map[uuid.UUID]uuid.UUID
	tokenContextCache       map[uuid.UUID]context.Context
	resourceContextProvider *ResourceContextProvider

	SubscriptionReader              billing_in.SubscriptionReader
	PlayerProfileReader             squad_in.PlayerProfileReader
	BillableOperationCommandHandler billing_in.BillableOperationCommandHandler
}

func NewRPCServer(c container.Container) *rpcServer {
	var subscriptionReader billing_in.SubscriptionReader

	err := c.Resolve(&subscriptionReader)

	if err != nil {
		slog.Error("Failed to resolve subscription reader", "error", err)
		return nil
	}

	var playerProfileReader squad_in.PlayerProfileReader

	err = c.Resolve(&playerProfileReader)

	if err != nil {
		slog.Error("Failed to resolve player profile reader", "error", err)
		return nil
	}

	var billableOperationCommandHandler billing_in.BillableOperationCommandHandler

	err = c.Resolve(&billableOperationCommandHandler)

	if err != nil {
		slog.Error("Failed to resolve billable operation command handler", "error", err)
		return nil
	}

	// c.
	return &rpcServer{
		Container:                       c,
		resourceContextProvider:         NewResourceContextProvider(&c),
		userTokenCache:                  make(map[uuid.UUID]map[uuid.UUID]iam_entities.RIDToken),
		tokenUserCache:                  make(map[uuid.UUID]uuid.UUID),
		tokenContextCache:               make(map[uuid.UUID]context.Context),
		SubscriptionReader:              subscriptionReader,
		PlayerProfileReader:             playerProfileReader,
		BillableOperationCommandHandler: billableOperationCommandHandler,
	}
}

func (s *rpcServer) GetContextWithUser(ctx context.Context, ridToken uuid.UUID) (context.Context, error) {
	userID, ok := s.tokenUserCache[ridToken]

	if ok {
		t := s.userTokenCache[userID][ridToken]
		if !t.IsExpired() {
			return s.tokenContextCache[ridToken], nil
		}
	}

	ctx, err := s.resourceContextProvider.GetVerifiedContext(ctx, ridToken)

	if err != nil {
		return nil, err
	}

	if s.userTokenCache[userID] != nil {

		for k, v := range s.userTokenCache[userID] {
			if v.IsExpired() {
				delete(s.userTokenCache[userID], k)
				delete(s.tokenUserCache, k)
				delete(s.tokenContextCache, k)
			}
		}
	}

	s.tokenContextCache[ridToken] = ctx

	return ctx, nil
}

// ValidateRID implements the gRPC endpoint to validate an RID.
func (s *rpcServer) ValidateRID(ctx context.Context, req *iam.ValidateRIDRequest) (*iam.ValidateRIDResponse, error) {
	if req.RidToken == "" {
		slog.Error("RID token is required")
		return nil, status.Error(codes.InvalidArgument, "RID token is required")
	}

	rid, err := uuid.Parse(req.RidToken)

	if err != nil {
		slog.Error("Invalid RID token (::1)")
		return nil, status.Error(codes.InvalidArgument, "Invalid RID token (::1)")
	}

	_, err = s.GetContextWithUser(ctx, rid)

	if err != nil {
		slog.Error("Invalid RID token (::2)")
		return nil, status.Error(codes.Unauthenticated, "Invalid RID token (::2)")
	}

	slog.Info("RID is valid")

	return &iam.ValidateRIDResponse{
		IsValid: true,
		Reason:  "RID is valid",
	}, nil
}

// GetPlayerProfile implements the gRPC endpoint to get a player profile.
func (s *rpcServer) GetPlayerProfile(ctx context.Context, req *squad.GetPlayerProfileRequest) (*squad.GetPlayerProfileResponse, error) {
	if req.PlayerId == "" {
		return nil, status.Error(codes.InvalidArgument, "Player ID is required")
	}

	// rid -> active profile (profile qual foi gerado o token)
	// se informado playerid, verifica se audiencia do token utilizado e client. (caso de async service no mm)
	// update: com audiencia correta no contexto, vai filtrar de acordo com o acesso concedido

	return &squad.GetPlayerProfileResponse{
		IsValid: true,
		Reason:  "Player profile retrieved successfully",
		PlayerProfile: &squad.PlayerProfile{
			Id:       req.PlayerId,
			Nickname: "SampleNickname",
		},
	}, nil
}

// GetSubscription implements the gRPC endpoint to get a subscription.
func (s *rpcServer) GetSubscription(ctx context.Context, req *billing.GetSubscriptionRequest) (*billing.GetSubscriptionResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}
	// Add logic here to retrieve the subscription according to your requirements.
	return &billing.GetSubscriptionResponse{
		IsValid: true,
		Reason:  "Subscription retrieved successfully",
		Subscription: &billing.Subscription{
			UserId: req.UserId,
			PlanId: "SamplePlanId",
		},
	}, nil
}

// ValidateOperation implements the gRPC endpoint to validate an operation.
func (s *rpcServer) ValidateOperation(ctx context.Context, req *billing.ValidateOperationRequest) (*billing.ValidateOperationResponse, error) {
	err := s.BillableOperationCommandHandler.Validate(ctx, billing_in.BillableOperationCommand{
		OperationID: billing_entities.BillableOperationKey(req.OperationId),
		UserID:      uuid.MustParse(req.UserId),
		Amount:      req.Amount,
		Args:        make(map[string]interface{}),
	})

	if err != nil {
		slog.Error("Failed to validate operation", "error", err)
		return nil, status.Error(codes.Internal, "Failed to validate operation")
	}

	return &billing.ValidateOperationResponse{
		IsValid: true,
		Reason:  "Operation is valid",
	}, nil
}

// ConfirmOperation implements the gRPC endpoint to confirm an operation.
func (s *rpcServer) ConfirmOperation(ctx context.Context, req *billing.ConfirmOperationRequest) (*billing.ConfirmOperationResponse, error) {
	result, err, validationErr := s.BillableOperationCommandHandler.Exec(ctx, billing_in.BillableOperationCommand{
		OperationID: billing_entities.BillableOperationKey(req.OperationId),
		UserID:      uuid.MustParse(req.UserId),
		Amount:      req.Amount,
		Args:        make(map[string]interface{}),
	})

	if validationErr != nil {
		slog.Error("Validation error during operation confirmation", "error", validationErr)
		return nil, status.Error(codes.InvalidArgument, "Validation error during operation confirmation")
	}

	if err != nil {
		slog.Error("Failed to confirm operation", "error", err)
		return nil, status.Error(codes.Internal, "Failed to confirm operation")
	}

	slog.Info("Operation confirmed successfully", "result", result)

	// Add logic here to confirm the operation according to your requirements.
	return &billing.ConfirmOperationResponse{
		Success: true,
		Reason:  "Operation confirmed successfully",
	}, nil
}

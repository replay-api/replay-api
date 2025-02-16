package replay_common

import billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"

const (
	OperationTypeUploadedReplayFiles billing_entities.BillableOperationKey = "UploadReplayFiles"
	OperationTypeStorageLimit        billing_entities.BillableOperationKey = "StorageLimit"
)

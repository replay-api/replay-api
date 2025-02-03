package squad_usecases

import (
	iam_out "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/ports/out"
	squad_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/ports/in"
	squad_out "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/ports/out"
)

type CreateSquadUseCase struct {
	SquadWriter squad_out.SquadWriter
	GroupWriter iam_out.GroupWriter // opcional create group: rule quantidade de grupos ou deixa infinito?
	// MembershipWriter iam_out.MembershipWriter
}

func NewCreateSquadUseCase(squadWriter squad_out.SquadWriter) *CreateSquadUseCase {
	return &CreateSquadUseCase{SquadWriter: squadWriter}
}

func (useCase *CreateSquadUseCase) Execute(squad *squad_in.CreateSquadCommand) (*squad_entities.Squad, error) {
	// TODO: verificar planos etc
	// TODO: consultar players

	// squad := squad_entities.NewSquad()

	// return useCase.SquadWriter.Create(squad) => create Profile with type: squad:public, squad:group (same for players, creatre profile with source=(player:public or player:group or player:private), key=slug) // alternativa=>> player:TenantAudience, player:ClientAudience, player:GroupAudience, player:UserAudience <= AudienceLevel

	return nil, nil
}

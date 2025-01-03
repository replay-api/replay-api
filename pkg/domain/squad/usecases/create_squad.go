package squad_usecases

import (
	squad_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/ports/in"
	squad_out "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/ports/out"
)

type CreateSquadUseCase struct {
	squadWriter squad_out.SquadWriter
}

func NewCreateSquadUseCase(squadWriter squad_out.SquadWriter) *CreateSquadUseCase {
	return &CreateSquadUseCase{squadWriter: squadWriter}
}

func (useCase *CreateSquadUseCase) Execute(squad *squad_in.CreateSquadCommand) (*squad_entities.Squad, error) {
	// TODO: verificar planos etc
	// TODO: consultar players

	// squad := squad_entities.NewSquad()

	// return useCase.squadWriter.Create(squad)

	return nil, nil
}

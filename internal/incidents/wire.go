//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build wireinject

package incidents

import (
	"github.com/google/wire"
)

func NewIncidentWire() *IncidentController {
	wire.Build(
		ControllerSet,
	)
	return &IncidentController{}
}

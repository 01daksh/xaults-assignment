package incidents

import (
	"xaults-assignment/internal/services"

	"github.com/google/wire"
)

var ControllerSet = wire.NewSet(
	NewIncidentController,
	NewIncidentService,
	NewIncidentRepository,
	services.NewServiceRepository,
)

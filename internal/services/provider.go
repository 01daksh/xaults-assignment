package services

import (
	"github.com/google/wire"
)

// depedency injection set for controllers
var ControllerSet = wire.NewSet(
	NewServiceController,
	NewServiceService,
	NewServiceRepository,

)

package activity

import (
	"svc-transaction/service"

	"github.com/sirupsen/logrus"
)

type Activity struct {
	logger *logrus.Logger

	service *service.Service
}

func NewActivity(
	logger *logrus.Logger,
	service *service.Service,
) *Activity {
	return &Activity{
		logger: logger,

		service: service,
	}
}

// GetActivities returns all Temporal activities for this service
func (api *Activity) GetActivities() []any {
	return []any{
		api.DebitAccount,
		api.CreditAccount,
		api.CompensateDebit,
	}
}

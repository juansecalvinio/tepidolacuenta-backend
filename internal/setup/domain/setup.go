package domain

import (
	branchDomain "juansecalvinio/tepidolacuenta/internal/branch/domain"
	restaurantDomain "juansecalvinio/tepidolacuenta/internal/restaurant/domain"
	tableDomain "juansecalvinio/tepidolacuenta/internal/table/domain"
)

// SetupRestaurantInput represents the data needed to create a restaurant with its first branch and tables
type SetupRestaurantInput struct {
	Name       string `json:"name" binding:"required,min=3,max=100"`
	CUIT       string `json:"cuit" binding:"required"`
	Address    string `json:"address" binding:"required,max=200"`
	TableCount int    `json:"tableCount" binding:"required,min=1,max=100"`
}

// SetupRestaurantOutput represents the complete result of the setup operation
type SetupRestaurantOutput struct {
	Restaurant *restaurantDomain.Restaurant `json:"restaurant"`
	Branch     *branchDomain.Branch         `json:"branch"`
	Tables     []*tableDomain.Table         `json:"tables"`
}

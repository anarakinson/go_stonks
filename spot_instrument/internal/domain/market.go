package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserRole int

var (
	UserRole_BASIC        UserRole = 0
	UserRole_PROFESSIONAL UserRole = 1
	UserRole_WHALE        UserRole = 2
)

type Market struct {
	ID             string
	Name           string
	Enabled        bool
	DeletedAt      *time.Time
	AvailableRoles []UserRole
}

func NewMarket(name string, enabled bool, available... UserRole) *Market {
	return &Market{
		ID:             uuid.New().String(),
		Name:           name,
		Enabled:        enabled,
		AvailableRoles: available,
	}
}

func (m *Market) Delete() {
	t := time.Now().UTC()
	m.DeletedAt = &t
}

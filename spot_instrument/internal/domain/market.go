package domain

import (
	"time"

	"github.com/google/uuid"
)

type Market struct {
	ID        string
	Name      string
	Enabled   bool
	DeletedAt *time.Time
}

func NewMarket(name string, enabled bool) *Market {
	return &Market{
		ID:      uuid.New().String(),
		Name:    name,
		Enabled: enabled,
	}
}

func (m *Market) Delete() {
	t := time.Now().UTC()
	m.DeletedAt = &t
}

package parking

import "fmt"

// ParkingSpot is satisfied by every concrete spot type.
type ParkingSpot interface {
	ID() string
	Type() SpotType
	IsFree() bool
	Assign(v Vehicle) error
	Vacate()
	CurrentVehicle() Vehicle // nil when free
}

// baseSpot holds the shared state and logic for all spot types.
type baseSpot struct {
	id        string
	spotType  SpotType
	available bool
	vehicle   Vehicle
}

func newBase(id string, t SpotType) baseSpot {
	return baseSpot{id: id, spotType: t, available: true}
}

func (s *baseSpot) ID() string             { return s.id }
func (s *baseSpot) Type() SpotType         { return s.spotType }
func (s *baseSpot) IsFree() bool           { return s.available }
func (s *baseSpot) CurrentVehicle() Vehicle { return s.vehicle }

func (s *baseSpot) Assign(v Vehicle) error {
	if !s.available {
		return fmt.Errorf("no spots available for %s", v.Type())
	}
	s.vehicle = v
	s.available = false
	return nil
}

func (s *baseSpot) Vacate() {
	s.vehicle = nil
	s.available = true
}

// Concrete spot types.

type CarSpot struct{ baseSpot }
type MotorcycleSpot struct{ baseSpot }
type DisabledSpot struct{ baseSpot }
type ElectricCarSpot struct{ baseSpot }
type LargeSpot struct{ baseSpot }

func NewCarSpot(id string) *CarSpot {
	return &CarSpot{newBase(id, SpotTypeCompact)}
}
func NewMotorcycleSpot(id string) *MotorcycleSpot {
	return &MotorcycleSpot{newBase(id, SpotTypeMotorcycle)}
}
func NewDisabledSpot(id string) *DisabledSpot {
	return &DisabledSpot{newBase(id, SpotTypeDisabled)}
}
func NewElectricCarSpot(id string) *ElectricCarSpot {
	return &ElectricCarSpot{newBase(id, SpotTypeElectricCar)}
}
func NewLargeSpot(id string) *LargeSpot {
	return &LargeSpot{newBase(id, SpotTypeLarge)}
}

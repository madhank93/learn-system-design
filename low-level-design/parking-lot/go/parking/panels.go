package parking

import (
	"fmt"
	"time"
)

// Clock abstracts time so tests can control it.
type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

// defaultClock is the production clock.
var defaultClock Clock = realClock{}

// EntryPanel issues parking tickets at the entrance gate.
type EntryPanel struct {
	id    string
	clock Clock
}

func NewEntryPanel(id string) *EntryPanel {
	return &EntryPanel{id: id, clock: defaultClock}
}

func (ep *EntryPanel) ID() string { return ep.id }

// SetClock overrides the clock — use in tests only.
func (ep *EntryPanel) SetClock(c Clock) { ep.clock = c }

// GetTicket finds an available floor+spot, assigns the vehicle, and returns a ticket.
func (ep *EntryPanel) GetTicket(v Vehicle) (*ParkingTicket, error) {
	lot := GetInstance()

	var floor *ParkingFloor
	for _, f := range lot.Floors() {
		if f.CanPark(v) {
			floor = f
			break
		}
	}
	if floor == nil {
		return nil, fmt.Errorf("parking unsupported for %s: lot is full or no compatible spots exist", v.Type())
	}

	spot := floor.AvailableSpot(v)
	if spot == nil {
		// CanPark passed but AvailableSpot returned nil — should not happen
		// with the fixed CanPark, but handled defensively.
		return nil, fmt.Errorf("no spots available for %s", v.Type())
	}

	ticket := newParkingTicket(v.Type(), v.Number(), floor.ID(), spot.ID(), ep.clock.Now())

	if err := spot.Assign(v); err != nil {
		return nil, err
	}
	return ticket, nil
}

// ExitPanel processes vehicle checkout at the exit gate.
type ExitPanel struct {
	id    string
	clock Clock
	cost  *HourlyCost
}

func NewExitPanel(id string) *ExitPanel {
	return &ExitPanel{id: id, clock: defaultClock, cost: newHourlyCost()}
}

func (xp *ExitPanel) ID() string { return xp.id }

// SetClock overrides the clock — use in tests only.
func (xp *ExitPanel) SetClock(c Clock) { xp.clock = c }

// Checkout vacates the spot, calculates the fee, and settles the ticket.
func (xp *ExitPanel) Checkout(ticket *ParkingTicket) error {
	endTime := xp.clock.Now()

	spot, err := GetInstance().VacateSpot(ticket.SpotID())
	if err != nil {
		return fmt.Errorf("checkout: %w", err)
	}

	rate, err := xp.cost.rate(spot.Type())
	if err != nil {
		return fmt.Errorf("checkout: %w", err)
	}

	ticket.settle(endTime, ticket.billedHours(endTime)*rate)
	return nil
}

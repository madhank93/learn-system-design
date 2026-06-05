package parking

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"time"
)

// ParkingTicket records a single parking session.
type ParkingTicket struct {
	id            string
	vehicleType   VehicleType
	vehicleNumber string
	floorID       string
	spotID        string
	startTime     time.Time
	endTime       time.Time
	amount        float64
}

func newTicketID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func newParkingTicket(vt VehicleType, number, floorID, spotID string, start time.Time) *ParkingTicket {
	return &ParkingTicket{
		id:            newTicketID(),
		vehicleType:   vt,
		vehicleNumber: number,
		floorID:       floorID,
		spotID:        spotID,
		startTime:     start,
	}
}

func (t *ParkingTicket) ID() string            { return t.id }
func (t *ParkingTicket) VehicleType() VehicleType { return t.vehicleType }
func (t *ParkingTicket) VehicleNumber() string  { return t.vehicleNumber }
func (t *ParkingTicket) FloorID() string        { return t.floorID }
func (t *ParkingTicket) SpotID() string         { return t.spotID }
func (t *ParkingTicket) StartTime() time.Time   { return t.startTime }
func (t *ParkingTicket) EndTime() time.Time     { return t.endTime }
func (t *ParkingTicket) Amount() float64        { return t.amount }

func (t *ParkingTicket) settle(end time.Time, amount float64) {
	t.endTime = end
	t.amount = amount
}

// billedHours returns the duration in whole hours (minimum 1).
func (t *ParkingTicket) billedHours(end time.Time) float64 {
	h := math.Round(end.Sub(t.startTime).Hours())
	if h < 1 {
		return 1
	}
	return h
}

func (t *ParkingTicket) String() string {
	return fmt.Sprintf(
		"Ticket{id=%.8s type=%s number=%s floor=%s spot=%s start=%s amount=%.2f}",
		t.id, t.vehicleType, t.vehicleNumber,
		t.floorID, t.spotID,
		t.startTime.Format("15:04"),
		t.amount,
	)
}

// HourlyCost maps spot types to hourly rates.
type HourlyCost struct {
	rates map[SpotType]float64
}

func newHourlyCost() *HourlyCost {
	return &HourlyCost{
		rates: map[SpotType]float64{
			SpotTypeMotorcycle:  15,
			SpotTypeDisabled:    15,
			SpotTypeCompact:     25,
			SpotTypeElectricCar: 35,
			SpotTypeLarge:       45,
		},
	}
}

func (h *HourlyCost) rate(st SpotType) (float64, error) {
	r, ok := h.rates[st]
	if !ok {
		return 0, fmt.Errorf("no rate configured for spot type %s", st)
	}
	return r, nil
}

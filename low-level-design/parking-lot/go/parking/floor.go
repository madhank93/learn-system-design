package parking

import "fmt"

// DisplayBoard shows available spot counts per type.
type DisplayBoard struct{}

func (d *DisplayBoard) Show(t SpotType, free int) {
	if free == 0 {
		fmt.Printf("[Display] %-12s FULL\n", t)
	} else {
		fmt.Printf("[Display] %-12s %d available\n", t, free)
	}
}

// ParkingFloor manages parking spots grouped by type.
type ParkingFloor struct {
	id    string
	spots map[SpotType][]ParkingSpot
	board *DisplayBoard
}

func NewParkingFloor(id string) *ParkingFloor {
	return &ParkingFloor{
		id: id,
		// BUG FIX: TS was missing SpotTypeDisabled here.
		spots: map[SpotType][]ParkingSpot{
			SpotTypeCompact:     {},
			SpotTypeMotorcycle:  {},
			SpotTypeLarge:       {},
			SpotTypeElectricCar: {},
			SpotTypeDisabled:    {},
		},
		board: &DisplayBoard{},
	}
}

func (f *ParkingFloor) ID() string                         { return f.id }
func (f *ParkingFloor) Spots() map[SpotType][]ParkingSpot  { return f.spots }

func (f *ParkingFloor) AddSpot(s ParkingSpot) {
	f.spots[s.Type()] = append(f.spots[s.Type()], s)
}

// CanPark returns true only when at least one spot of the correct type is FREE.
// BUG FIX: TS checked len(spots) > 0, which returned true even when all spots
// were occupied, causing the subsequent GetAvailableSpot to return nil and
// producing a misleading "no spots available" error instead of blocking entry.
func (f *ParkingFloor) CanPark(v Vehicle) bool {
	for _, s := range f.spots[vehicleToSpotType(v.Type())] {
		if s.IsFree() {
			return true
		}
	}
	return false
}

// AvailableSpot returns the first free spot for the vehicle type, or nil.
func (f *ParkingFloor) AvailableSpot(v Vehicle) ParkingSpot {
	for _, s := range f.spots[vehicleToSpotType(v.Type())] {
		if s.IsFree() {
			return s
		}
	}
	return nil
}

// InUseSpots returns all occupied spots for the given vehicle type.
// BUG FIX: TS used `!spot.isSpotFree` (a function reference, always truthy)
// instead of `!spot.isSpotFree()` — so it always returned an empty list.
func (f *ParkingFloor) InUseSpots(vt VehicleType) []ParkingSpot {
	var result []ParkingSpot
	for _, s := range f.spots[vehicleToSpotType(vt)] {
		if !s.IsFree() {
			result = append(result, s)
		}
	}
	return result
}

func (f *ParkingFloor) ShowDisplayBoard() {
	for t, spots := range f.spots {
		free := 0
		for _, s := range spots {
			if s.IsFree() {
				free++
			}
		}
		f.board.Show(t, free)
	}
}

package parking

import (
	"fmt"
	"sync"
)

var (
	lotInstance *ParkingLot
	lotOnce     sync.Once
)

// ParkingLot is the Singleton root of the system.
type ParkingLot struct {
	floors      []*ParkingFloor
	entryPanels []*EntryPanel
	exitPanels  []*ExitPanel
}

// GetInstance returns the single ParkingLot instance.
func GetInstance() *ParkingLot {
	lotOnce.Do(func() {
		lotInstance = &ParkingLot{}
	})
	return lotInstance
}

// ResetForTest resets the singleton. Call at the start of each test that
// uses GetInstance() to avoid state leaking between tests.
func ResetForTest() {
	lotOnce = sync.Once{}
	lotInstance = nil
}

func (pl *ParkingLot) Floors() []*ParkingFloor      { return pl.floors }
func (pl *ParkingLot) EntryPanels() []*EntryPanel   { return pl.entryPanels }
func (pl *ParkingLot) ExitPanels() []*ExitPanel     { return pl.exitPanels }

// VacateSpot scans ALL floors and ALL spot-type buckets for the given spot ID.
// BUG FIX: the TS implementation had `return vacatedSpot` inside the inner
// loop body, causing it to exit after checking only the FIRST spot-type bucket
// on the FIRST floor — it would return nil if the spot was not in that bucket,
// even when the spot existed elsewhere in the lot.
func (pl *ParkingLot) VacateSpot(spotID string) (ParkingSpot, error) {
	for _, floor := range pl.floors {
		for _, spots := range floor.Spots() {
			for _, spot := range spots {
				if spot.ID() == spotID {
					spot.Vacate()
					return spot, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("spot %q not found in parking lot", spotID)
}

// Admin manages the parking lot structure.
type Admin struct {
	username string
	password string
}

func NewAdmin(username, password string) *Admin {
	return &Admin{username: username, password: password}
}

func (a *Admin) AddFloor(f *ParkingFloor) error {
	lot := GetInstance()
	for _, existing := range lot.floors {
		if existing.ID() == f.ID() {
			return fmt.Errorf("floor %s already exists", f.ID())
		}
	}
	lot.floors = append(lot.floors, f)
	return nil
}

func (a *Admin) AddSpot(floorID string, s ParkingSpot) error {
	lot := GetInstance()
	for _, f := range lot.floors {
		if f.ID() != floorID {
			continue
		}
		for _, spots := range f.spots {
			for _, existing := range spots {
				if existing.ID() == s.ID() {
					return fmt.Errorf("spot %s already exists on floor %s", s.ID(), floorID)
				}
			}
		}
		f.AddSpot(s)
		return nil
	}
	return fmt.Errorf("floor %s not found", floorID)
}

func (a *Admin) AddEntryPanel(ep *EntryPanel) error {
	lot := GetInstance()
	for _, existing := range lot.entryPanels {
		if existing.ID() == ep.ID() {
			return fmt.Errorf("entry panel %s already exists", ep.ID())
		}
	}
	lot.entryPanels = append(lot.entryPanels, ep)
	return nil
}

func (a *Admin) AddExitPanel(xp *ExitPanel) error {
	lot := GetInstance()
	for _, existing := range lot.exitPanels {
		if existing.ID() == xp.ID() {
			return fmt.Errorf("exit panel %s already exists", xp.ID())
		}
	}
	lot.exitPanels = append(lot.exitPanels, xp)
	return nil
}

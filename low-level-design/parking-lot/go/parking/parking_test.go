package parking_test

import (
	"testing"
	"time"

	"parkinglot/parking"
)

// mockClock lets tests control time precisely.
type mockClock struct{ t time.Time }

func (c *mockClock) Now() time.Time           { return c.t }
func (c *mockClock) Advance(d time.Duration)  { c.t = c.t.Add(d) }

// setup creates a fresh singleton + admin for each test.
func setup(t *testing.T) *parking.Admin {
	t.Helper()
	parking.ResetForTest()
	return parking.NewAdmin("admin", "pass")
}

// ── Vehicle ──────────────────────────────────────────────────

func TestVehicle_TypeAndNumber(t *testing.T) {
	car := parking.NewCar("TN12345")
	bike := parking.NewMotorcycle("TN10005")

	if car.Type() != parking.VehicleTypeCar {
		t.Errorf("want %s, got %s", parking.VehicleTypeCar, car.Type())
	}
	if car.Number() != "TN12345" {
		t.Errorf("want TN12345, got %s", car.Number())
	}
	if bike.Type() != parking.VehicleTypeMotorcycle {
		t.Errorf("want %s, got %s", parking.VehicleTypeMotorcycle, bike.Type())
	}
	if bike.Number() != "TN10005" {
		t.Errorf("want TN10005, got %s", bike.Number())
	}
}

// ── ParkingSpot ───────────────────────────────────────────────

func TestParkingSpot_TypeAndID(t *testing.T) {
	cs := parking.NewCarSpot("2F001")
	ms := parking.NewMotorcycleSpot("1F010")

	if cs.Type() != parking.SpotTypeCompact {
		t.Errorf("want %s, got %s", parking.SpotTypeCompact, cs.Type())
	}
	if ms.Type() != parking.SpotTypeMotorcycle {
		t.Errorf("want %s, got %s", parking.SpotTypeMotorcycle, ms.Type())
	}
	if cs.ID() != "2F001" {
		t.Errorf("want 2F001, got %s", cs.ID())
	}
}

func TestParkingSpot_StartsAsFree(t *testing.T) {
	s := parking.NewCarSpot("S1")
	if !s.IsFree() {
		t.Error("new spot should be free")
	}
}

func TestParkingSpot_BecomesOccupiedAfterAssign(t *testing.T) {
	s := parking.NewCarSpot("S1")
	if err := s.Assign(parking.NewCar("TN123")); err != nil {
		t.Fatal(err)
	}
	if s.IsFree() {
		t.Error("spot should be occupied after Assign")
	}
}

func TestParkingSpot_FreeAgainAfterVacate(t *testing.T) {
	s := parking.NewCarSpot("S1")
	_ = s.Assign(parking.NewCar("TN123"))
	s.Vacate()
	if !s.IsFree() {
		t.Error("spot should be free after Vacate")
	}
}

func TestParkingSpot_AssignToOccupiedReturnsError(t *testing.T) {
	s := parking.NewCarSpot("S1")
	car := parking.NewCar("TN123")
	_ = s.Assign(car)

	err := s.Assign(car)
	if err == nil {
		t.Fatal("expected error assigning to occupied spot")
	}
}

func TestParkingSpot_CurrentVehicle(t *testing.T) {
	s := parking.NewCarSpot("S1")
	car := parking.NewCar("TN12345")
	_ = s.Assign(car)

	v := s.CurrentVehicle()
	if v == nil {
		t.Fatal("expected non-nil vehicle")
	}
	if v.Number() != "TN12345" {
		t.Errorf("want TN12345, got %s", v.Number())
	}
	if v.Type() != parking.VehicleTypeCar {
		t.Errorf("want Car, got %s", v.Type())
	}
}

func TestParkingSpot_VacateClearsVehicle(t *testing.T) {
	s := parking.NewCarSpot("S1")
	_ = s.Assign(parking.NewCar("TN123"))
	s.Vacate()
	if s.CurrentVehicle() != nil {
		t.Error("CurrentVehicle should be nil after Vacate")
	}
}

// ── ParkingFloor ──────────────────────────────────────────────

func TestParkingFloor_ID(t *testing.T) {
	f := parking.NewParkingFloor("F001")
	if f.ID() != "F001" {
		t.Errorf("want F001, got %s", f.ID())
	}
}

func TestParkingFloor_HasAllFiveSpotTypes(t *testing.T) {
	f := parking.NewParkingFloor("F001")
	want := []parking.SpotType{
		parking.SpotTypeCompact,
		parking.SpotTypeMotorcycle,
		parking.SpotTypeLarge,
		parking.SpotTypeElectricCar,
		parking.SpotTypeDisabled, // was missing in TS — bug fix
	}
	for _, st := range want {
		if _, ok := f.Spots()[st]; !ok {
			t.Errorf("floor missing spot type %s", st)
		}
	}
}

func TestParkingFloor_InitialSpotsEmpty(t *testing.T) {
	f := parking.NewParkingFloor("F001")
	for st, spots := range f.Spots() {
		if len(spots) != 0 {
			t.Errorf("spot type %s should start empty, got %d", st, len(spots))
		}
	}
}

func TestParkingFloor_AddSpot(t *testing.T) {
	f := parking.NewParkingFloor("F001")
	f.AddSpot(parking.NewCarSpot("CS1"))
	if len(f.Spots()[parking.SpotTypeCompact]) != 1 {
		t.Error("expected 1 compact spot after AddSpot")
	}
}

func TestParkingFloor_AvailableSpotNoneWhenEmpty(t *testing.T) {
	f := parking.NewParkingFloor("F001")
	if f.AvailableSpot(parking.NewCar("TN1")) != nil {
		t.Error("expected nil when no spots exist")
	}
}

func TestParkingFloor_AvailableSpotReturnsFirstFree(t *testing.T) {
	f := parking.NewParkingFloor("F001")
	s1 := parking.NewCarSpot("CS1")
	s2 := parking.NewCarSpot("CS2")
	f.AddSpot(s1)
	f.AddSpot(s2)

	car := parking.NewCar("TN1")
	got := f.AvailableSpot(car)
	if got == nil || got.ID() != "CS1" {
		t.Errorf("want CS1, got %v", got)
	}

	_ = s1.Assign(car)
	got = f.AvailableSpot(parking.NewCar("TN2"))
	if got == nil || got.ID() != "CS2" {
		t.Errorf("want CS2 after CS1 occupied, got %v", got)
	}
}

func TestParkingFloor_CanPark_ReturnsTrueOnlyWhenFreeSpotExists(t *testing.T) {
	f := parking.NewParkingFloor("F001")
	car := parking.NewCar("TN1")

	// No spots at all — should return false
	if f.CanPark(car) {
		t.Error("CanPark should be false with no spots")
	}

	s := parking.NewCarSpot("CS1")
	f.AddSpot(s)

	// Spot exists and is free — should return true
	if !f.CanPark(car) {
		t.Error("CanPark should be true when free spot exists")
	}

	// BUG FIX: occupy the only spot — CanPark must now return false
	// (TS returned true here because it only checked len(spots) > 0)
	_ = s.Assign(car)
	if f.CanPark(parking.NewCar("TN2")) {
		t.Error("CanPark should be false when all spots are occupied (bug fix)")
	}
}

func TestParkingFloor_InUseSpots(t *testing.T) {
	f := parking.NewParkingFloor("F001")
	s1 := parking.NewCarSpot("CS1")
	s2 := parking.NewCarSpot("CS2")
	f.AddSpot(s1)
	f.AddSpot(s2)

	// BUG FIX: TS used !spot.isSpotFree (fn reference) — always returned empty
	_ = s1.Assign(parking.NewCar("TN1"))
	occupied := f.InUseSpots(parking.VehicleTypeCar)
	if len(occupied) != 1 {
		t.Errorf("expected 1 in-use spot, got %d", len(occupied))
	}
	if occupied[0].ID() != "CS1" {
		t.Errorf("expected CS1 in-use, got %s", occupied[0].ID())
	}
}

// ── EntryPanel ────────────────────────────────────────────────

func TestEntryPanel_IssuesTicket(t *testing.T) {
	admin := setup(t)
	floor := parking.NewParkingFloor("F001")
	spot := parking.NewCarSpot("PS1001")
	_ = admin.AddFloor(floor)
	floor.AddSpot(spot)

	ep := parking.NewEntryPanel("ENT001")
	ticket, err := ep.GetTicket(parking.NewCar("TN112345"))
	if err != nil {
		t.Fatal(err)
	}

	if ticket.VehicleType() != parking.VehicleTypeCar {
		t.Errorf("want Car, got %s", ticket.VehicleType())
	}
	if ticket.VehicleNumber() != "TN112345" {
		t.Errorf("want TN112345, got %s", ticket.VehicleNumber())
	}
	if ticket.SpotID() != "PS1001" {
		t.Errorf("want PS1001, got %s", ticket.SpotID())
	}
	if ticket.FloorID() != "F001" {
		t.Errorf("want F001, got %s", ticket.FloorID())
	}
	if ticket.Amount() != 0 {
		t.Errorf("amount should be 0 before checkout, got %.2f", ticket.Amount())
	}
	if spot.IsFree() {
		t.Error("spot should be occupied after ticket issued")
	}
}

func TestEntryPanel_SpotOccupiedAfterTicket(t *testing.T) {
	admin := setup(t)
	floor := parking.NewParkingFloor("F001")
	spot := parking.NewCarSpot("CS1")
	_ = admin.AddFloor(floor)
	floor.AddSpot(spot)

	ep := parking.NewEntryPanel("ENT001")
	_, _ = ep.GetTicket(parking.NewCar("TN1"))

	if spot.IsFree() {
		t.Error("spot should be occupied after parking")
	}
}

func TestEntryPanel_ErrorWhenNoCompatibleFloor(t *testing.T) {
	admin := setup(t)
	floor := parking.NewParkingFloor("F001")
	_ = admin.AddFloor(floor)
	floor.AddSpot(parking.NewMotorcycleSpot("MS1"))

	ep := parking.NewEntryPanel("ENT001")
	// Try to park a Car when only motorcycle spots exist
	_, err := ep.GetTicket(parking.NewCar("TN1"))
	if err == nil {
		t.Fatal("expected error when no compatible floor")
	}
}

func TestEntryPanel_ErrorWhenAllSpotsOccupied(t *testing.T) {
	admin := setup(t)
	floor := parking.NewParkingFloor("F001")
	spot := parking.NewCarSpot("CS1")
	_ = admin.AddFloor(floor)
	floor.AddSpot(spot)

	ep := parking.NewEntryPanel("ENT001")

	// First car parks fine
	_, err := ep.GetTicket(parking.NewCar("TN1"))
	if err != nil {
		t.Fatal(err)
	}

	// Second car should fail — CanPark now returns false when spot occupied (bug fix)
	_, err = ep.GetTicket(parking.NewCar("TN2"))
	if err == nil {
		t.Fatal("expected error when all spots occupied")
	}
}

// ── ExitPanel ─────────────────────────────────────────────────

func TestExitPanel_Checkout_2Hours_Compact(t *testing.T) {
	admin := setup(t)
	floor := parking.NewParkingFloor("F001")
	spot := parking.NewCarSpot("CS1")
	_ = admin.AddFloor(floor)
	floor.AddSpot(spot)

	clock := &mockClock{t: time.Date(2023, 1, 13, 10, 0, 0, 0, time.UTC)}

	ep := parking.NewEntryPanel("ENT001")
	ep.SetClock(clock)

	ticket, err := ep.GetTicket(parking.NewCar("TN112345"))
	if err != nil {
		t.Fatal(err)
	}

	clock.Advance(2 * time.Hour) // +2 hours

	xp := parking.NewExitPanel("EXT001")
	xp.SetClock(clock)

	if err := xp.Checkout(ticket); err != nil {
		t.Fatal(err)
	}

	// Compact = ₹25/hr × 2hr = ₹50
	if ticket.Amount() != 50 {
		t.Errorf("want 50, got %.2f", ticket.Amount())
	}
	if !spot.IsFree() {
		t.Error("spot should be free after checkout")
	}
}

func TestExitPanel_Checkout_MinimumOneHour(t *testing.T) {
	admin := setup(t)
	floor := parking.NewParkingFloor("F001")
	floor.AddSpot(parking.NewMotorcycleSpot("MS1"))
	_ = admin.AddFloor(floor)

	clock := &mockClock{t: time.Date(2023, 1, 13, 10, 0, 0, 0, time.UTC)}

	ep := parking.NewEntryPanel("ENT001")
	ep.SetClock(clock)
	ticket, _ := ep.GetTicket(parking.NewMotorcycle("TN1"))

	clock.Advance(20 * time.Minute) // less than 1 hour

	xp := parking.NewExitPanel("EXT001")
	xp.SetClock(clock)
	_ = xp.Checkout(ticket)

	// Motorcycle = ₹15/hr × min 1hr = ₹15
	if ticket.Amount() != 15 {
		t.Errorf("want 15 (minimum 1 hour), got %.2f", ticket.Amount())
	}
}

func TestExitPanel_Checkout_3Hours_Large(t *testing.T) {
	admin := setup(t)
	floor := parking.NewParkingFloor("F001")
	floor.AddSpot(parking.NewLargeSpot("LS1"))
	_ = admin.AddFloor(floor)

	clock := &mockClock{t: time.Date(2023, 1, 13, 8, 0, 0, 0, time.UTC)}

	ep := parking.NewEntryPanel("ENT001")
	ep.SetClock(clock)
	ticket, _ := ep.GetTicket(parking.NewTruck("TN9999"))

	clock.Advance(3 * time.Hour)

	xp := parking.NewExitPanel("EXT001")
	xp.SetClock(clock)
	_ = xp.Checkout(ticket)

	// Large = ₹45/hr × 3hr = ₹135
	if ticket.Amount() != 135 {
		t.Errorf("want 135, got %.2f", ticket.Amount())
	}
}

// ── VacateSpot bug fix ────────────────────────────────────────

func TestVacateSpot_FindsSpotAcrossAllFloors(t *testing.T) {
	// BUG FIX: TS vacateParkingSpot returned after checking only the first
	// spot-type bucket on the first floor. This test places the target spot
	// on the second floor to prove the fixed implementation scans all floors.
	admin := setup(t)

	floor1 := parking.NewParkingFloor("F001")
	floor2 := parking.NewParkingFloor("F002")
	_ = admin.AddFloor(floor1)
	_ = admin.AddFloor(floor2)

	// Add a car spot only on floor 2
	floor1.AddSpot(parking.NewMotorcycleSpot("MS1")) // floor1 has only motorcycle spots
	floor2.AddSpot(parking.NewCarSpot("CS1"))

	clock := &mockClock{t: time.Now()}
	ep := parking.NewEntryPanel("ENT001")
	ep.SetClock(clock)

	ticket, err := ep.GetTicket(parking.NewCar("TN1"))
	if err != nil {
		t.Fatal(err)
	}
	if ticket.FloorID() != "F002" {
		t.Errorf("expected car parked on F002, got %s", ticket.FloorID())
	}

	clock.Advance(time.Hour)
	xp := parking.NewExitPanel("EXT001")
	xp.SetClock(clock)

	if err := xp.Checkout(ticket); err != nil {
		t.Errorf("checkout failed (vacate bug): %v", err)
	}
}

// ── Admin ─────────────────────────────────────────────────────

func TestAdmin_DuplicateFloorRejected(t *testing.T) {
	admin := setup(t)
	f := parking.NewParkingFloor("F001")
	if err := admin.AddFloor(f); err != nil {
		t.Fatal(err)
	}
	if err := admin.AddFloor(f); err == nil {
		t.Error("expected error adding duplicate floor")
	}
}

func TestAdmin_DuplicateSpotRejected(t *testing.T) {
	admin := setup(t)
	f := parking.NewParkingFloor("F001")
	_ = admin.AddFloor(f)
	s := parking.NewCarSpot("CS1")

	if err := admin.AddSpot("F001", s); err != nil {
		t.Fatal(err)
	}
	if err := admin.AddSpot("F001", s); err == nil {
		t.Error("expected error adding duplicate spot")
	}
}

func TestAdmin_AddSpotToUnknownFloorErrors(t *testing.T) {
	admin := setup(t)
	if err := admin.AddSpot("UNKNOWN", parking.NewCarSpot("CS1")); err == nil {
		t.Error("expected error for unknown floor")
	}
}

# Parking Lot — Go

Go implementation of the parking lot LLD, with all bugs from the TypeScript version fixed.

## Requirements

Managed via [mise](https://mise.jdx.dev/) — Go version pinned in `/.mise.toml` at the repo root.

| Tool | Version |
|------|---------|
| Go   | 1.25.1  |

```bash
mise install   # installs Go
```

## Run tests

```bash
go test ./parking/ -v
```

Expected output — 27 tests:

```
=== RUN   TestVehicle_TypeAndNumber
--- PASS: TestVehicle_TypeAndNumber
=== RUN   TestParkingSpot_TypeAndID
--- PASS: TestParkingSpot_TypeAndID
...
PASS
ok  parkinglot/parking
```

## Run the demo

```bash
go run .
```

Sample output:

```
=== Display Board ===
[Display] Compact      2 available
[Display] Motorcycle   2 available
[Display] Disabled     1 available
...

Car1 ticket: Ticket{id=3cdf4d74 type=Car number=TN23890 floor=PF1001 spot=CPS1001 start=08:16 end=- amount=0.00}
Car1 checked out. Amount: ₹25
Truck checked out. Amount: ₹45
```

## Project structure

```
go/
├── go.mod
├── main.go                  Demo — wires up the full system end-to-end
└── parking/
    ├── types.go             VehicleType + SpotType enums, vehicle→spot mapping
    ├── vehicle.go           Vehicle interface + Car/Motorcycle/Truck/ElectricCar/Van
    ├── spot.go              ParkingSpot interface + CarSpot/MotorcycleSpot/DisabledSpot/ElectricCarSpot/LargeSpot
    ├── floor.go             ParkingFloor + DisplayBoard
    ├── lot.go               ParkingLot (Singleton) + Admin
    ├── ticket.go            ParkingTicket + HourlyCost
    ├── panels.go            EntryPanel + ExitPanel + Clock interface
    └── parking_test.go      27 tests
```

## Bugs fixed from TypeScript version

| Bug | TypeScript | Go fix |
|-----|-----------|--------|
| `vacateParkingSpot` exits after first spot-type bucket | `return vacatedSpot` inside inner loop | `return` only when spot found, continues scanning all floors |
| `canPark` returns true when all spots occupied | `spots.length > 0` | `some(s => s.IsFree())` |
| `getInUseSpotID` always returns empty | `!spot.isSpotFree` (fn reference) | `!s.IsFree()` (method call) |
| `DisabledSpot` type missing from floor init | Only 4 types initialized | All 5 types initialized |

## Pricing

| Spot Type  | Rate (per hour) |
|------------|-----------------|
| Motorcycle | ₹15             |
| Disabled   | ₹15             |
| Compact    | ₹25             |
| ElectricCar| ₹35             |
| Large      | ₹45             |

Minimum charge: 1 hour.

# Parking Lot — TypeScript

Low-level design implementation of a parking lot system in TypeScript.

## Requirements

Managed via [mise](https://mise.jdx.dev/) — versions pinned in `/.mise.toml` at the repo root.

| Tool | Version |
|------|---------|
| Node | 22.22.3 |
| pnpm | 11.5.2 |

```bash
mise install   # installs Node + pnpm
```

## Install dependencies

```bash
pnpm install
```

## Run tests

```bash
pnpm test
```

## Run the demo

```bash
pnpm run index
```

## Project structure

```
src/
├── account/
│   ├── account.ts           Abstract account base class
│   ├── admin.ts             Admin — manages floors, spots, panels
│   └── parking.attendant.ts Parking attendant account
├── parking-spot/
│   ├── parking.spot.ts      Abstract ParkingSpot base class
│   ├── parking.spot.type.ts ParkingSpotType enum
│   ├── car.spot.ts          Compact spot (cars)
│   ├── motorcycle.spot.ts   Motorcycle spot
│   ├── disabled.spot.ts     Disabled spot
│   └── electric.car.spot.ts Electric car spot
├── vehicles/
│   ├── vehicle.ts           Abstract Vehicle base class
│   ├── vehicle.type.ts      VehicleType enum
│   ├── car.ts               Car
│   ├── motor.cycle.ts       Motorcycle
│   ├── truck.ts             Truck
│   ├── electric.ts          Electric car
│   └── van.ts               (unused — reserved for Van type)
├── parking.lot.ts           ParkingLot singleton
├── parking.floor.ts         ParkingFloor — manages spots per type
├── entry.panel.ts           EntryPanel — issues parking tickets
├── exit.panel.ts            ExitPanel — processes checkout + billing
├── parking.ticket.ts        ParkingTicket value object
├── parking.display.board.ts DisplayBoard — shows free spot counts
└── hourly.cost.ts           HourlyCost — pricing by spot type

test/
├── vehicle.spec.ts
├── parking.spot.spec.ts
├── parking.floor.spec.ts
├── entry.panel.spec.ts
├── entry.panel.negative.spec.ts
└── exit.panel.spec.ts
```

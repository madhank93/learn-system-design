import { describe, expect, it } from "vitest";
import Admin from "../src/account/admin";
import { EntryPanel } from "../src/entry.panel";
import { MotorcycleSpot } from "../src/parking-spot/motorcycle.spot";
import { ParkingFloor } from "../src/parking.floor";
import { Car } from "../src/vehicles/car";
import { MotorCycle } from "../src/vehicles/motor.cycle";

// TODO: Keep tests related to entry panel in a single file once test level isolation implemented
describe("Entry panel negative tests cases", () => {
  it("Verify its throws an error when parking lot doesn't have spot for that vehicle type", () => {
    const car = new Car("TN112345");
    const entryPanel = new EntryPanel("Panel001");
    const floor = new ParkingFloor("F1001");
    const bikeSpot = new MotorcycleSpot("BK1001");

    const admin = new Admin("admin", "12345");
    admin.addParkingFloor(floor);

    floor
      .getListOfParkingSpots()
      .get(bikeSpot.getParkingSpotType())
      ?.push(bikeSpot);

    expect(() => entryPanel.getParkingTicket(car)).toThrowError(
      "Parking is unsupported for this Car type"
    );
  });

  it("Verify its throws an error when parking floor doesn't have spot", () => {
    const bike = new MotorCycle("TN112345");
    const entryPanel = new EntryPanel("Panel001");
    const floor = new ParkingFloor("F1001");
    const bikeSpot = new MotorcycleSpot("BK1001");

    const admin = new Admin("admin", "12345");
    admin.addParkingFloor(floor);

    floor
      .getListOfParkingSpots()
      .get(bikeSpot.getParkingSpotType())
      ?.push(bikeSpot);

    entryPanel.getParkingTicket(bike);

    expect(() => entryPanel.getParkingTicket(bike)).toThrowError(
      "No spots are available for Motorcycle"
    );
  });
});

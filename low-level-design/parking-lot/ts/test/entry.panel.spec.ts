import { describe, expect, it } from "vitest";
import Admin from "../src/account/admin";
import { EntryPanel } from "../src/entry.panel";
import { CarSpot } from "../src/parking-spot/car.spot";
import { ParkingFloor } from "../src/parking.floor";
import { Car } from "../src/vehicles/car";

describe("Entry panel tests cases", () => {
  it("Verify able to generate parking ticket", () => {
    const car = new Car("TN112345");
    const entryPanel = new EntryPanel("Panel001");
    const firstFloor = new ParkingFloor("F001");
    const carSpot = new CarSpot("PS1001");

    const admin = new Admin("admin", "12345");
    admin.addParkingFloor(firstFloor);

    firstFloor
      .getListOfParkingSpots()
      .get(carSpot.getParkingSpotType())
      ?.push(carSpot);

    const ticket = entryPanel.getParkingTicket(car);

    const ticketExpected = {
      vehicleType: "Car",
      vehicleRegisterNumber: "TN112345",
      parkingSpotID: "PS1001",
      parkingFloorID: "F001",
      endTime: undefined,
      amount: undefined,
    };
    expect(ticket).toMatchObject(ticketExpected);
  });
});

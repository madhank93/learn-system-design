import { describe, it } from "vitest";
import Admin from "../src/account/admin";
import { EntryPanel } from "../src/entry.panel";
import { ExitPanel } from "../src/exit.panel";

import { CarSpot } from "../src/parking-spot/car.spot";
import { ParkingFloor } from "../src/parking.floor";
import { Car } from "../src/vehicles/car";

describe("Exit panel tests cases", () => {
  it("Verify check out process of the vehicle", () => {
    const car = new Car("TN112345");
    const entryPanel = new EntryPanel("Panel001");
    const exitPanel = new ExitPanel();
    const firstFloor = new ParkingFloor("F001");
    const carSpot = new CarSpot("PS1001");

    const admin = new Admin("admin", "12345");
    admin.addParkingFloor(firstFloor);

    firstFloor
      .getListOfParkingSpots()
      .get(carSpot.getParkingSpotType())
      ?.push(carSpot);

    let ticket = entryPanel.getParkingTicket(car);

    ticket = exitPanel.checkout(ticket);

    console.log("********", ticket);
  });
});

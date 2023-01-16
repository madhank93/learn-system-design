import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import Admin from "../src/account/admin";
import { EntryPanel } from "../src/entry.panel";
import { ExitPanel } from "../src/exit.panel";
import { CarSpot } from "../src/parking-spot/car.spot";
import { ParkingFloor } from "../src/parking.floor";
import { Car } from "../src/vehicles/car";

describe("Exit panel tests cases", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("Verify check out process of the vehicle", () => {
    const car = new Car("TN112345");
    const entryPanel = new EntryPanel("ENT001");
    const exitPanel = new ExitPanel("EXT001");
    const firstFloor = new ParkingFloor("F001");
    const carSpot = new CarSpot("PS1001");
    const startDate = new Date(2023, 0, 13, 10);

    vi.setSystemTime(startDate);

    const admin = new Admin("admin", "12345");
    admin.addParkingFloor(firstFloor);
    firstFloor
      .getListOfParkingSpots()
      .get(carSpot.getParkingSpotType())
      ?.push(carSpot);
    let ticket = entryPanel.getParkingTicket(car);

    vi.advanceTimersByTime(3600000);
    ticket = exitPanel.checkout(ticket);

    expect(ticket.getAmount()).toEqual(25);
  });
});

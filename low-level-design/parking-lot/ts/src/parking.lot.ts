import { EntryPanel } from "./entry.panel";
import { ExitPanel } from "./exit.panel";
import { ParkingFloor } from "./parking.floor";

export class ParkingLot {
  private static instance: ParkingLot;
  private parkingFloor: Array<ParkingFloor>;
  private entryPanels: Array<EntryPanel>;
  private exitPanels: Array<ExitPanel>;

  private constructor() {
    this.parkingFloor = new Array<ParkingFloor>();
    this.entryPanels = new Array<EntryPanel>();
    this.exitPanels = new Array<ExitPanel>();
  }

  public static getInstance() {
    if (!ParkingLot.instance) {
      return (this.instance = new ParkingLot());
    }
    return this.instance;
  }

  public vacateParkingSpot(parkingSpotID: string) {
    for (const floor of this.parkingFloor) {
      for (const spots of floor.getListOfParkingSpots().values()) {
        const spot = spots.find((s) => s.getParkingSpotID() === parkingSpotID);
        if (spot) {
          spot.vacateVehicleFromSpot();
          return spot;
        }
      }
    }
    return undefined;
  }

  public getListOfParkingFloor() {
    return this.parkingFloor;
  }

  public getListOfEntryPanel() {
    return this.entryPanels;
  }

  public getListOfExitPanel() {
    return this.exitPanels;
  }
}

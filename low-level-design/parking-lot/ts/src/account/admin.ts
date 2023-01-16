import { EntryPanel } from "../entry.panel";
import { ExitPanel } from "../exit.panel";
import { ParkingSpot } from "../parking-spot/parking.spot";
import { ParkingFloor } from "../parking.floor";
import { ParkingLot } from "../parking.lot";
import { Account } from "./account";

export default class Admin extends Account {
  constructor(username: string, password: string) {
    super(username, password);
  }

  public addParkingFloor(parkingFloor: ParkingFloor): boolean {
    const floor = ParkingLot.getInstance()
      .getListOfParkingFloor()
      .find(
        (pf) => pf.getParkingFloorID() === parkingFloor.getParkingFloorID()
      );
    if (floor === undefined) {
      ParkingLot.getInstance().getListOfParkingFloor().push(parkingFloor);
      return true;
    }
    return false;
  }

  public addParkingSpot(parkingFloorID: string, parkingSpot: ParkingSpot) {
    const floor = ParkingLot.getInstance()
      .getListOfParkingFloor()
      .find((pf) => pf.getParkingFloorID() === parkingFloorID);

    if (floor === undefined) {
      throw new Error("Invalid floor");
    }

    const spot = floor
      .getListOfParkingSpots()
      .get(parkingSpot.getParkingSpotType())
      ?.find(
        (spot) => spot.getParkingSpotID() === parkingSpot.getParkingSpotID()
      );
    if (spot !== undefined) {
      return;
    }
    floor
      .getListOfParkingSpots()
      .get(parkingSpot.getParkingSpotType())
      ?.push(parkingSpot);
  }

  public addEntrancePanel(entryPanel: EntryPanel) {
    const listOfEntryPanel = ParkingLot.getInstance().getListOfEntryPanel();
    const result = listOfEntryPanel.find(
      (panel) => panel.getEntryPanelID() === entryPanel.getEntryPanelID()
    );

    if (result === undefined) {
      listOfEntryPanel.push(entryPanel);
      return true;
    }
    return false;
  }

  public addExitPanel(exitPanel: ExitPanel) {
    const listOfExistPanel = ParkingLot.getInstance().getListOfExitPanel();
    const result = listOfExistPanel.find(
      (panel) => panel.getExitPanelID() === exitPanel.getExitPanelID()
    );

    if (result === undefined) {
      listOfExistPanel.push(exitPanel);
      return true;
    }
    return false;
  }
}

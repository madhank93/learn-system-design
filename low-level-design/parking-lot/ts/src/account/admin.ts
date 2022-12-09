import ParkingSpot from "../parking-spot/parking.spot";
import { ParkingSpotType } from "../parking-spot/parking.spot.type";
import ParkingFloor from "../parking.floor";
import Account from "./account";

export default class Admin extends Account {
  constructor(username: string, password: string) {
    super(username, password);
  }

  public addParkingFloor(parkingFloor: ParkingFloor): boolean {
    return false;
  }

  public addParkingSpot(
    parkingFloorID: string,
    listOfParkingSpot: Array<ParkingSpot>
  ) {
    // this.parkingSpots.set(ParkingSpotType.Compact, listOfParkingSpot);
  }
}

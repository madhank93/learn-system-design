import ParkingFloor from "../parking.floor";
import Account from "./account";

export default class Admin extends Account {
  constructor(username: string, password: string) {
    super(username, password);
  }

  public addParkingFloor(parkingFloor: ParkingFloor): boolean {
    return false;
  }
}

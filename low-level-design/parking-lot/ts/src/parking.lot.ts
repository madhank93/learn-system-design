import ParkingFloor from "./parking.floor";

export default class ParkingLot {
  private static INSTANCE: ParkingLot = new ParkingLot();
  private parkingFloor: Array<ParkingFloor>;
  private constructor() {
    this.parkingFloor = new Array<ParkingFloor>();
  }

  public getParkingFloor(parkingFloorID: string) {
    ParkingLot.INSTANCE.parkingFloor;
  }
}

import ParkingFloor from "./parking.floor";
import Vehicle from "./vehicles/vehicle";

export default class ParkingLot {
  private static instance: ParkingLot;
  private parkingFloor: Array<ParkingFloor>;

  private constructor() {
    this.parkingFloor = new Array<ParkingFloor>();
  }

  public static getInstance() {
    if (!ParkingLot.instance) {
      return (this.instance = new ParkingLot());
    }
    return this.instance;
  }

  public getParkingSpot(vehicle: Vehicle): ParkingFloor | undefined {
    return this.parkingFloor.find((floor) => floor.canPark(vehicle));
  }

  public getListOfParkingFloor() {
    return this.parkingFloor;
  }
}

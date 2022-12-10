import ParkingFloor from "./parking.floor";
import { VehicleType } from "./vehicles/vehicle.type";

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

  public getParkingSpot(vehicleType: VehicleType): ParkingFloor | undefined {
    return this.parkingFloor.find((floor) => floor.canPark(vehicleType));
  }

  public getListOfParkingFloor() {
    return this.parkingFloor;
  }
}

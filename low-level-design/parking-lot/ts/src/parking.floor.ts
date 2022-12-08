import ParkingSpot from "./parking-spot/parking.spot";
import { VehicleType } from "./vehicles/vehicle.type";

export default class ParkingFloor {
  private parkingSpots: Map<VehicleType, Array<ParkingSpot>>;
  private parkingSpotCapacity: number;
  private parkingFloorID: string;

  constructor(parkingSpotCapacity: number, parkingFloorID: string) {
    this.parkingSpotCapacity = parkingSpotCapacity;
    this.parkingSpots = new Map<VehicleType, Array<ParkingSpot>>();
    this.parkingFloorID = parkingFloorID;
  }

  public addParkingSpot(parkingFloorID: string) {}
}

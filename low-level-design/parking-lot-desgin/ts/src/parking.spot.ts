import { ParkingSpotType } from "./parking.spot.type";
import Vehicle from "./vehicle";

export default abstract class ParkingSpot {
  private id: string;
  private spotAvailable: boolean;
  private vehicle: Vehicle | null;
  private parkingSpotType: ParkingSpotType;

  constructor(id: string, parkingSpotType: ParkingSpotType) {
    this.id = id;
    this.spotAvailable = true;
    this.vehicle = null;
    this.parkingSpotType = parkingSpotType;
  }

  public isFree(): boolean {
    return false;
  }

  public assignVehicle(vehicle: Vehicle) {
    this.vehicle = vehicle;
    this.spotAvailable = false;
  }

  public removeVehicle() {
    this.vehicle = null;
    this.spotAvailable = true;
  }
}

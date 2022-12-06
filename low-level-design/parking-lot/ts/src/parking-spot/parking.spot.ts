import { ParkingSpotType } from "./parking.spot.type";
import Vehicle from "../vehicles/vehicle";

export default abstract class ParkingSpot {
  private id: string;
  private isSpotAvailable: boolean;
  private vehicle: Vehicle | null;
  private parkingSpotType: ParkingSpotType;

  constructor(id: string, parkingSpotType: ParkingSpotType) {
    this.id = id;
    this.isSpotAvailable = true;
    this.vehicle = null;
    this.parkingSpotType = parkingSpotType;
  }

  public isFree(): boolean {
    return false;
  }

  public assignVehicle(vehicle: Vehicle) {
    if (!this.isSpotAvailable) {
      throw new Error("Spot is unavailable");
    }
    this.vehicle = vehicle;
    this.isSpotAvailable = false;
  }

  public removeVehicle() {
    this.vehicle = null;
    this.isSpotAvailable = true;
  }
}

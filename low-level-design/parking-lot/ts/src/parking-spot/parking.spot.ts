import { ParkingSpotType } from "./parking.spot.type";
import Vehicle from "../vehicles/vehicle";

export default abstract class ParkingSpot {
  private parkingSpotId: string;
  private isSpotAvailable: boolean;
  private vehicle: Vehicle | null;
  private parkingSpotType: ParkingSpotType;

  constructor(parkingSpotId: string, parkingSpotType: ParkingSpotType) {
    this.isSpotAvailable = true;
    this.parkingSpotId = parkingSpotId;
    this.vehicle = null;
    this.parkingSpotType = parkingSpotType;
  }

  public isSpotFree(): boolean {
    return this.isSpotAvailable;
  }

  public assignVehicleToSpot(vehicle: Vehicle) {
    if (!this.isSpotAvailable) {
      throw new Error("Spot is unavailable");
    }
    this.vehicle = vehicle;
    this.isSpotAvailable = false;
  }

  public removeVehicle(vehicle: Vehicle) {
    this.vehicle = null;
    this.isSpotAvailable = true;
  }
}

import { ParkingSpotType } from "./parking.spot.type";
import Vehicle from "../vehicles/vehicle";

export default abstract class ParkingSpot {
  private parkingSpotID: string;
  private isSpotAvailable: boolean;
  private vehicle: Vehicle | null;
  private parkingSpotType: ParkingSpotType;

  constructor(parkingSpotId: string, parkingSpotType: ParkingSpotType) {
    this.isSpotAvailable = true;
    this.parkingSpotID = parkingSpotId;
    this.vehicle = null;
    this.parkingSpotType = parkingSpotType;
  }

  public isSpotFree(): boolean {
    return this.isSpotAvailable;
  }

  public getParkingSpotType() {
    return this.parkingSpotType;
  }

  public getParkingSpotID() {
    return this.parkingSpotID;
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

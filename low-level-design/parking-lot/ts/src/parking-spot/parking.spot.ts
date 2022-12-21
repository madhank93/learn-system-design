import { ParkingSpotType } from "./parking.spot.type";
import { Vehicle } from "../vehicles/vehicle";
import { SpotOccupiedException } from "../errors/spot.occupied.exception";

export abstract class ParkingSpot {
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

  public getVehicleDetails() {
    return this.vehicle;
  }

  public assignVehicleToSpot(vehicle: Vehicle) {
    if (!this.isSpotAvailable) {
      throw new SpotOccupiedException(vehicle.getVehicleType());
    }
    this.vehicle = vehicle;
    this.isSpotAvailable = false;
    // console.log(vehicle, this.isSpotAvailable);
  }

  public vacateVehicleFromSpot() {
    this.vehicle = null;
    this.isSpotAvailable = true;
  }
}

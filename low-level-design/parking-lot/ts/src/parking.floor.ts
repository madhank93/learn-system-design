import ParkingSpot from "./parking-spot/parking.spot";
import { ParkingSpotType } from "./parking-spot/parking.spot.type";
import { VehicleType } from "./vehicles/vehicle.type";

export default class ParkingFloor {
  private parkingSpots: Map<ParkingSpotType, Array<ParkingSpot>> = new Map();
  private parkingFloorID: string;

  constructor(parkingFloorID: string) {
    this.parkingSpots.set(ParkingSpotType.Compact, new Array<ParkingSpot>());
    this.parkingFloorID = parkingFloorID;
  }

  public getParkingFloorID() {
    return this.parkingFloorID;
  }

  public getParkingSpots() {
    return this.parkingSpots;
  }

  private getSpotTypeForVehicle(vehicleType: VehicleType): ParkingSpotType {
    switch (vehicleType) {
      case VehicleType.Car:
        return ParkingSpotType.Compact;
      case VehicleType.Motorcycle:
        return ParkingSpotType.Motorcycle;
      default:
        return ParkingSpotType.Large;
    }
  }

  public canPark(vehicleType: VehicleType) {
    this.getSpotTypeForVehicle(vehicleType);
  }
}

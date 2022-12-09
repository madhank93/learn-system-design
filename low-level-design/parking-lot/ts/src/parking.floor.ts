import ParkingSpot from "./parking-spot/parking.spot";
import { ParkingSpotType } from "./parking-spot/parking.spot.type";
import Vehicle from "./vehicles/vehicle";
import { VehicleType } from "./vehicles/vehicle.type";

export default class ParkingFloor {
  private parkingSpots: Map<ParkingSpotType, Array<ParkingSpot>>;
  private parkingFloorID: string;

  constructor(
    parkingFloorID: string,
    parkingSpots: Map<ParkingSpotType, Array<ParkingSpot>>
  ) {
    this.parkingSpots = parkingSpots;
    this.parkingFloorID = parkingFloorID;
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
}

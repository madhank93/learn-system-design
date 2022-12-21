import ParkingSpot from "./parking-spot/parking.spot";
import { ParkingSpotType } from "./parking-spot/parking.spot.type";
import Vehicle from "./vehicles/vehicle";
import { VehicleType } from "./vehicles/vehicle.type";

export default class ParkingFloor {
  private parkingSpots: Map<ParkingSpotType, Array<ParkingSpot>> = new Map();
  private parkingFloorID: string;

  constructor(parkingFloorID: string) {
    this.parkingSpots.set(ParkingSpotType.Compact, new Array<ParkingSpot>());
    this.parkingSpots.set(ParkingSpotType.Motorcycle, new Array<ParkingSpot>());
    this.parkingSpots.set(ParkingSpotType.Large, new Array<ParkingSpot>());
    this.parkingSpots.set(ParkingSpotType.Electric, new Array<ParkingSpot>());
    this.parkingFloorID = parkingFloorID;
  }

  public getParkingFloorID() {
    return this.parkingFloorID;
  }

  public getListOfParkingSpots() {
    return this.parkingSpots;
  }

  public getAvailableSpot(vehicle: Vehicle) {
    return this.parkingSpots
      .get(this.getSpotTypeForVehicle(vehicle.getVehicleType()))
      ?.find((spot) => spot.isSpotFree());
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

  public getInUseSpotID(vehicleType: VehicleType) {
    return this.parkingSpots
      .get(this.getSpotTypeForVehicle(vehicleType))
      ?.filter((spot) => !spot.isSpotFree)
      .map((occSpot) => occSpot);
  }

  public canPark(vehicle: Vehicle) {
    const parkingSpotType = this.getSpotTypeForVehicle(
      vehicle.getVehicleType()
    );

    if (this.parkingSpots.get(parkingSpotType)?.length) {
      return true;
    }

    return false;

    // const freeSpot = this.parkingSpots
    //   .get(parkingSpotType)
    //   ?.find((spot) => spot.isSpotFree());

    // if (freeSpot === undefined) {
    //   throw new Error("Parking spot for this vehicle type is unavailable");
    // }
    // return freeSpot;
  }
}

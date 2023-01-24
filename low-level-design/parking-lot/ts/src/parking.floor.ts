import { DisplayBoard } from "./parking.display.board";
import { ParkingSpot } from "./parking-spot/parking.spot";
import { ParkingSpotType } from "./parking-spot/parking.spot.type";
import { Vehicle } from "./vehicles/vehicle";
import { VehicleType } from "./vehicles/vehicle.type";

export class ParkingFloor {
  private parkingSpots: Map<ParkingSpotType, Array<ParkingSpot>> = new Map();
  private parkingFloorID: string;
  private displayBoard: DisplayBoard;

  constructor(parkingFloorID: string) {
    this.parkingSpots.set(ParkingSpotType.Compact, new Array<ParkingSpot>());
    this.parkingSpots.set(ParkingSpotType.Motorcycle, new Array<ParkingSpot>());
    this.parkingSpots.set(ParkingSpotType.Large, new Array<ParkingSpot>());
    this.parkingSpots.set(
      ParkingSpotType.ElectricCar,
      new Array<ParkingSpot>()
    );
    this.parkingFloorID = parkingFloorID;
    this.displayBoard = new DisplayBoard();
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

  public showDisplayBoard() {
    for (let spotType of this.parkingSpots.keys()) {
      const freeSpots =
        this.parkingSpots.get(spotType)?.filter((spot) => spot.isSpotFree()) ??
        [];
      this.displayBoard.displayMessage(spotType, freeSpots.length);
    }
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
  }
}

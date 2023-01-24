import { ParkingSpot } from "./parking.spot";
import { ParkingSpotType } from "./parking.spot.type";

export class ElectricCarSpot extends ParkingSpot {
  constructor(id: string) {
    super(id, ParkingSpotType.ElectricCar);
  }
}

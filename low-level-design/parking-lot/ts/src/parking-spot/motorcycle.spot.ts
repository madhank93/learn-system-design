import { ParkingSpot } from "./parking.spot";
import { ParkingSpotType } from "./parking.spot.type";

export class MotorcycleSpot extends ParkingSpot {
  constructor(id: string) {
    super(id, ParkingSpotType.Motorcycle);
  }
}

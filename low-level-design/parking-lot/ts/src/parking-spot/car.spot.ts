import { ParkingSpot } from "./parking.spot";
import { ParkingSpotType } from "./parking.spot.type";

export class CarSpot extends ParkingSpot {
  constructor(id: string) {
    super(id, ParkingSpotType.Compact);
  }
}

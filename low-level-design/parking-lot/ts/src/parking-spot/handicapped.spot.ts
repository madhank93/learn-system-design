import { ParkingSpot } from "./parking.spot";
import { ParkingSpotType } from "./parking.spot.type";

export class HandicappedSpot extends ParkingSpot {
  constructor(id: string) {
    super(id, ParkingSpotType.Handicapped);
  }
}

import { ParkingSpot } from "./parking.spot";
import { ParkingSpotType } from "./parking.spot.type";

export class DisabledSpot extends ParkingSpot {
  constructor(id: string) {
    super(id, ParkingSpotType.Disabled);
  }
}

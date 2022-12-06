import HandicappedSpot from "./parking-spot/handicapped.spot";
import ParkingSpot from "./parking-spot/parking.spot";

export default class ParkingFloor {
  //private id: string;
  private parkingSpots: Array<ParkingSpot>;

  constructor(parkingSpotCapacity: number) {
    this.parkingSpots = new Array<ParkingSpot>(parkingSpotCapacity);
  }

  addValue() {}
}

const test = new ParkingFloor(4);
test.addValue();

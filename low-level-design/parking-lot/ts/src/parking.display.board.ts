import { ParkingSpotType } from "./parking-spot/parking.spot.type";

export class DisplayBoard {
  public displayMessage(parkingSpotType: ParkingSpotType, freeSpots: number) {
    let message = `${parkingSpotType} spots free: ${freeSpots}`;
    console.log(message);
  }
}

import { ParkingSpotType } from "./parking-spot/parking.spot.type";

export class HourlyCost {
  private hourlyCosts: Map<ParkingSpotType, number> = new Map();

  constructor() {
    this.hourlyCosts.set(ParkingSpotType.Compact, 25);
    this.hourlyCosts.set(ParkingSpotType.Disabled, 15);
    this.hourlyCosts.set(ParkingSpotType.ElectricCar, 35);
    this.hourlyCosts.set(ParkingSpotType.Large, 45);
    this.hourlyCosts.set(ParkingSpotType.Motorcycle, 15);
  }

  public getCost(parkingSpotType: ParkingSpotType) {
    if (this.hourlyCosts.has(parkingSpotType)) {
      return this.hourlyCosts.get(parkingSpotType)!;
    } else {
      throw new Error(
        `Hourly cost is not associated with ${parkingSpotType} type`
      );
    }
  }
}

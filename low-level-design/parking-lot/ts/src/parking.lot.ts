import ParkingFloor from "./parking.floor";
import Vehicle from "./vehicles/vehicle";

export default class ParkingLot {
  private static instance: ParkingLot;
  private parkingFloor: Array<ParkingFloor>;

  private constructor() {
    this.parkingFloor = new Array<ParkingFloor>();
  }

  public static getInstance() {
    if (!ParkingLot.instance) {
      return (this.instance = new ParkingLot());
    }
    return this.instance;
  }

  public vacateParkingSpot(parkingSpotID: string) {
    for (let floor of this.parkingFloor) {
      for (let spots of floor.getListOfParkingSpots().values()) {
        spots.find((spot) => {
          if (spot.getParkingSpotID() === parkingSpotID) {
            console.log(spot);
            spot.vacateVehicleFromSpot();
            console.log(spot);
          }
        });
      }
    }
  }

  public getListOfParkingFloor() {
    return this.parkingFloor;
  }
}

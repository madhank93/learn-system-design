import { HourlyCost } from "./hourly.cost";
import { ParkingSpotType } from "./parking-spot/parking.spot.type";
import { ParkingLot } from "./parking.lot";
import { ParkingTicket } from "./parking.ticket";
export class ExitPanel {
  public checkout(parkingTicket: ParkingTicket) {
    const parkingSpotID = parkingTicket.getParkingSpotID();
    const totalDurationInHours = this.calculateDuration(parkingTicket);
    const vacatedSpot =
      ParkingLot.getInstance().vacateParkingSpot(parkingSpotID);

    if (vacatedSpot === undefined) {
      throw new Error("Unable to find the Parking Spot for the given ID");
    }

    const totalAmount = this.calculatePrice(
      vacatedSpot.getParkingSpotType(),
      totalDurationInHours
    );

    parkingTicket.setAmount(totalAmount);
    console.log(parkingTicket);
  }

  private calculatePrice(parkingSpotType: ParkingSpotType, duration: number) {
    return duration * new HourlyCost().getCost(parkingSpotType);
  }

  private calculateDuration(parkingTicket: ParkingTicket) {
    const endTime = parkingTicket.setEndTime().getEndTime();
    return Math.round(
      Math.abs(parkingTicket.getStartTime().valueOf() - endTime.valueOf()) /
        36e5
    );
  }
}

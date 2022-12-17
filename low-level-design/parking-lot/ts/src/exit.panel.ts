import ParkingSpot from "./parking-spot/parking.spot";
import ParkingLot from "./parking.lot";
import ParkingTicket from "./parking.ticket";
export default class ExitPanel {
  private parkingTicket!: ParkingTicket;

  public checkout(parkingTicket: ParkingTicket) {
    const parkingSpotID = this.parkingTicket.getParkingSpotID();
    ParkingLot.getInstance().getParkingSpot();
  }
}

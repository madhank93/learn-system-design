import ParkingLot from "./parking.lot";
import ParkingTicket from "./parking.ticket";
export default class ExitPanel {
  public checkout(parkingTicket: ParkingTicket) {
    const parkingSpotID = parkingTicket.getParkingSpotID();
    ParkingLot.getInstance().vacateParkingSpot(parkingSpotID);
  }
}

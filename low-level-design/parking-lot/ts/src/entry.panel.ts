import ParkingLot from "./parking.lot";
import ParkingTicket from "./parking.ticket";
import Vehicle from "./vehicles/vehicle";

export default class EntryPanel {
  private entryPanelID: string;

  constructor(entryPanelID: string) {
    this.entryPanelID = entryPanelID;
  }

  public getParkingTicket(vehicle: Vehicle) {
    // this.generateParkingTicket(
    //   vehicle.getVehicleRegisterNumber(),
    //   parkingSpotID
    // );
  }

  private generateParkingTicket(
    vehicleRegisterNumber: string,
    parkingSpotID: string
  ) {
    // const parkingTicket = new ParkingTicket();
  }
}

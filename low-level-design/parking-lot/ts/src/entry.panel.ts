import SpotOccupiedException from "./errors/spot.occupied.exception";
import ParkingLot from "./parking.lot";
import ParkingTicket from "./parking.ticket";
import Vehicle from "./vehicles/vehicle";

export default class EntryPanel {
  private entryPanelID: string;

  constructor(entryPanelID: string) {
    this.entryPanelID = entryPanelID;
  }

  public getParkingTicket(vehicle: Vehicle) {
    const parkingFloor = ParkingLot.getInstance()
      .getListOfParkingFloor()
      .find((floor) => floor.canPark(vehicle));

    const spot = parkingFloor?.getAvailableSpot(vehicle);

    if (spot === undefined) {
      throw new SpotOccupiedException(vehicle.getVehicleType());
    }

    const ticket = this.generateParkingTicket(
      vehicle.getVehicleRegisterNumber(),
      spot.getParkingSpotID()
    );

    console.log("ticket: ", ticket);
  }

  private generateParkingTicket(
    vehicleRegisterNumber: string,
    parkingSpotID: string
  ) {
    return new ParkingTicket(vehicleRegisterNumber, parkingSpotID).setStartTime(
      Date.now()
    );
  }
}

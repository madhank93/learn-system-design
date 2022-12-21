import { SpotOccupiedException } from "./errors/spot.occupied.exception";
import { ParkingLot } from "./parking.lot";
import { ParkingTicket } from "./parking.ticket";
import { Vehicle } from "./vehicles/vehicle";

export class EntryPanel {
  private entryPanelID: string;

  constructor(entryPanelID: string) {
    this.entryPanelID = entryPanelID;
  }

  public getParkingTicket(vehicle: Vehicle) {
    const parkingFloor = ParkingLot.getInstance()
      .getListOfParkingFloor()
      .find((floor) => floor.canPark(vehicle));

    if (parkingFloor === undefined) {
      throw new Error(
        `Parking is unsupported for this ${vehicle.getVehicleType()} type`
      );
    }

    const spot = parkingFloor.getAvailableSpot(vehicle);

    if (spot === undefined) {
      throw new SpotOccupiedException(vehicle.getVehicleType());
    }

    const ticket = this.generateParkingTicket(
      vehicle.getVehicleRegisterNumber(),
      parkingFloor.getParkingFloorID(),
      spot.getParkingSpotID()
    );

    spot.assignVehicleToSpot(vehicle);

    console.log("ticket: ", ticket);
    return ticket;
  }

  private generateParkingTicket(
    vehicleRegisterNumber: string,
    parkingFloorID: string,
    parkingSpotID: string
  ) {
    return new ParkingTicket(
      vehicleRegisterNumber,
      parkingFloorID,
      parkingSpotID
    ).setStartTime(Date.now());
  }
}

import { ParkingLot } from "./parking.lot";
import { ParkingTicket } from "./parking.ticket";
import { Vehicle } from "./vehicles/vehicle";

export class EntryPanel {
  private entryPanelID: string;

  constructor(entryPanelID: string) {
    this.entryPanelID = entryPanelID;
  }

  public getEntryPanelID() {
    return this.entryPanelID;
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
      throw new Error(`No spots are available for ${vehicle.getVehicleType()}`);
    }

    const ticket = this.generateParkingTicket(
      vehicle,
      parkingFloor.getParkingFloorID(),
      spot.getParkingSpotID()
    );

    spot.assignVehicleToSpot(vehicle);
    return ticket;
  }

  private generateParkingTicket(
    vehicle: Vehicle,
    parkingFloorID: string,
    parkingSpotID: string
  ) {
    return new ParkingTicket(
      vehicle.getVehicleType(),
      vehicle.getVehicleRegisterNumber(),
      parkingFloorID,
      parkingSpotID
    ).setStartTime();
  }
}

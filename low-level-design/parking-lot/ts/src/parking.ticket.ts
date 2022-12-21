import crypto from "node:crypto";

export default class ParkingTicket {
  private parkingTicketID: string;
  private vehicleRegisterNumber: string;
  private parkingSpotID: string;
  private parkingFloorID: string;
  private startTime!: number;
  private endTime!: number;

  constructor(
    vehicleRegisterNumber: string,
    parkingFloorID: string,
    parkingSpotID: string
  ) {
    this.vehicleRegisterNumber = vehicleRegisterNumber;
    this.parkingFloorID = parkingFloorID;
    this.parkingSpotID = parkingSpotID;
    this.parkingTicketID = crypto.randomUUID();
  }

  public setStartTime(currentDateTime: number) {
    this.startTime = currentDateTime;
    return this;
  }

  public setEndTime(currentDateTime: number) {
    this.endTime = currentDateTime;
    return this;
  }

  public getParkingSpotID() {
    return this.parkingSpotID;
  }

  public getParkingFloorID() {
    return this.parkingFloorID;
  }

  public getVehicleRegisterNumber() {
    return this.vehicleRegisterNumber;
  }
}

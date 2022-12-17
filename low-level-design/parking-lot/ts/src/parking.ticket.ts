import crypto from "node:crypto";

export default class ParkingTicket {
  private parkingTicketID: string;
  private vehicleRegisterNumber: string;
  private parkingSpotID: string;
  private startTime!: number;
  private endTime!: number;

  constructor(vehicleRegisterNumber: string, parkingSpotID: string) {
    this.vehicleRegisterNumber = vehicleRegisterNumber;
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
}

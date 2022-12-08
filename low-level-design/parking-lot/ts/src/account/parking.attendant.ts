import Account from "./account";

export default class ParkingAttendant extends Account {
  constructor(username: string, password: string) {
    super(username, password);
  }

  public processTicket(): boolean {
    return false;
  }
}

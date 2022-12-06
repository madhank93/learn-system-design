import Account from "./account";

export default class Admin extends Account {
  constructor(username: string, password: string) {
    super(username, password);
  }

  public addParkingFloor(): boolean {
    return false;
  }
}

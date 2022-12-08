export default abstract class Account {
  private username: string;
  private password: string;

  constructor(username: string, password: string) {
    this.username = username;
    this.password = password;
  }

  public resetPassword(newPassword: string) {
    this.password = newPassword;
  }
}

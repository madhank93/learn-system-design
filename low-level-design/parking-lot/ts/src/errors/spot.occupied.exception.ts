export default class SpotOccupiedException extends Error {
  constructor(message: string = "Spot is occupied") {
    super(message);
  }
}

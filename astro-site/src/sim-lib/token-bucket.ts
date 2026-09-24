/**
 * Pure token-bucket logic, split from the React view (TokenBucket).
 * Tokens refill at a steady rate up to a capacity; each request spends one.
 */

/** Refill `tokens` for `dtMs` elapsed at `rate` tokens/sec, clamped to capacity. */
export function refill(tokens: number, capacity: number, rate: number, dtMs: number): number {
  return Math.min(capacity, tokens + (rate * dtMs) / 1000);
}

/** Canonical engine — drives tests and is reusable outside React. */
export class TokenBucket {
  tokens: number;
  constructor(
    public capacity: number,
    public rate: number,
    tokens: number = capacity,
  ) {
    this.tokens = tokens;
  }

  refill(dtMs: number): void {
    this.tokens = refill(this.tokens, this.capacity, this.rate, dtMs);
  }

  /** Spend `n` tokens if available; returns whether the request is allowed. */
  tryConsume(n = 1): boolean {
    if (this.tokens >= n) {
      this.tokens -= n;
      return true;
    }
    return false;
  }
}

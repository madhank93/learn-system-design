import { describe, it, expect } from 'vitest';
import { refill, TokenBucket } from './token-bucket';

describe('refill', () => {
  it('adds rate * dt tokens', () => {
    expect(refill(0, 10, 2, 1000)).toBe(2);
    expect(refill(0, 10, 2, 500)).toBe(1);
  });
  it('clamps at capacity', () => {
    expect(refill(9, 10, 5, 1000)).toBe(10);
  });
});

describe('TokenBucket', () => {
  it('starts full by default', () => {
    expect(new TokenBucket(10, 2).tokens).toBe(10);
  });
  it('allows a burst up to capacity then throttles', () => {
    const b = new TokenBucket(3, 1, 3);
    expect(b.tryConsume()).toBe(true);
    expect(b.tryConsume()).toBe(true);
    expect(b.tryConsume()).toBe(true);
    expect(b.tryConsume()).toBe(false); // drained
  });
  it('recovers at the refill rate', () => {
    const b = new TokenBucket(5, 2, 0);
    expect(b.tryConsume()).toBe(false);
    b.refill(1000); // +2 tokens
    expect(b.tryConsume()).toBe(true);
    expect(b.tryConsume()).toBe(true);
    expect(b.tryConsume()).toBe(false);
  });
});

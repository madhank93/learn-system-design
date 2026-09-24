import { describe, it, expect } from 'vitest';
import { recordCall, shouldTrip, CircuitBreaker, DEFAULT_CB } from './circuit-breaker';

describe('recordCall', () => {
  it('keeps only the last `window` entries', () => {
    let win: boolean[] = [];
    for (let i = 0; i < 12; i++) win = recordCall(win, true, 10);
    expect(win.length).toBe(10);
  });
});

describe('shouldTrip', () => {
  it('waits for minCalls before evaluating', () => {
    expect(shouldTrip([false, false, false], 5, 0.5)).toBe(false);
  });
  it('trips above the threshold', () => {
    expect(shouldTrip([false, false, false, true, true], 5, 0.5)).toBe(true);
  });
  it('stays closed at or below the threshold', () => {
    expect(shouldTrip([false, false, true, true, true], 5, 0.5)).toBe(false);
  });
});

describe('CircuitBreaker', () => {
  it('trips CLOSED → OPEN once failures cross the threshold', () => {
    const cb = new CircuitBreaker(DEFAULT_CB);
    for (let i = 0; i < 5; i++) cb.request(false, 0);
    expect(cb.state).toBe('OPEN');
  });
  it('fails fast (block) while OPEN', () => {
    const cb = new CircuitBreaker(DEFAULT_CB);
    for (let i = 0; i < 5; i++) cb.request(false, 0);
    expect(cb.request(true, 100)).toBe('block');
  });
  it('OPEN → HALF_OPEN after cooldown elapses', () => {
    const cb = new CircuitBreaker(DEFAULT_CB);
    for (let i = 0; i < 5; i++) cb.request(false, 0);
    expect(cb.tick(DEFAULT_CB.openMs - 1)).toBe(false);
    expect(cb.tick(DEFAULT_CB.openMs)).toBe(true);
    expect(cb.state).toBe('HALF_OPEN');
  });
  it('successful probe closes the breaker and resets the window', () => {
    const cb = new CircuitBreaker(DEFAULT_CB);
    for (let i = 0; i < 5; i++) cb.request(false, 0);
    cb.tick(DEFAULT_CB.openMs);
    expect(cb.request(true, DEFAULT_CB.openMs + 1)).toBe('probe-ok');
    expect(cb.state).toBe('CLOSED');
    expect(cb.win).toEqual([]);
  });
  it('failed probe re-opens the breaker', () => {
    const cb = new CircuitBreaker(DEFAULT_CB);
    for (let i = 0; i < 5; i++) cb.request(false, 0);
    cb.tick(DEFAULT_CB.openMs);
    expect(cb.request(false, DEFAULT_CB.openMs + 1)).toBe('probe-fail');
    expect(cb.state).toBe('OPEN');
  });
});

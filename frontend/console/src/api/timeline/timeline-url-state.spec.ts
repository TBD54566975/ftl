import { describe, expect, test } from 'vitest';
import { NicerURLSearchParams, TimelineUrlState } from './timeline-url-state';

describe('timeline url state', () => {
  test('empty', () => {
    const params = new NicerURLSearchParams();
    const timelineUrlState = new TimelineUrlState(params);
    expect(timelineUrlState.getSearchParams().toString()).toEqual(params.toString());
  });

  test('modules', () => {
    const params = new NicerURLSearchParams({
      modules: 'time,echo',
    });
    const timelineUrlState = new TimelineUrlState(params);
    expect(timelineUrlState.getSearchParams().toString()).toEqual('modules=time,echo');
  });

  test('log level', () => {
    const params = new NicerURLSearchParams({
      log: 'info',
    });
    const timelineUrlState = new TimelineUrlState(params);
    expect(timelineUrlState.getSearchParams().toString()).toEqual('log=info');
  });
});

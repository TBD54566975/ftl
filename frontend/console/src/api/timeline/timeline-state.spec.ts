import { describe, expect, test } from 'vitest'
import { getSearchParams, newTimelineState, NicerURLSearchParams, TimelineState } from './timeline-state'

const testCases = [
  {
    description: 'empty',
    params: new NicerURLSearchParams(),
    expected: '',
  },
  {
    description: 'tailing (default state)',
    params: new NicerURLSearchParams({ tail: 'live' }),
    expected: '',
  },
  {
    description: 'modules',
    params: new NicerURLSearchParams({ modules: 'time,echo' }),
    expected: 'modules=time,echo',
  },
  {
    description: 'log level',
    params: new NicerURLSearchParams({ log: 'info' }),
    expected: 'log=info',
  },
  {
    description: 'time range',
    params: new NicerURLSearchParams({
      after: '2021-09-01T00:00:00.000Z',
      before: '2021-09-02T00:00:00.000Z',
    }),
    expected: 'after=2021-09-01T00:00:00.000Z&before=2021-09-02T00:00:00.000Z',
  },
  {
    description: 'tailing paused',
    params: new NicerURLSearchParams({ tail: '1', paused: '1' }),
    expected: 'tail=paused',
  },
  {
    description: 'tailing with time range (incompatible settings--tail/paused should be ignored)',
    params: new NicerURLSearchParams({
      tail: 'paused',
      after: '2021-09-01T00:00:00.000Z',
      before: '2021-09-02T00:00:00.000Z',
    }),
    expected: 'after=2021-09-01T00:00:00.000Z&before=2021-09-02T00:00:00.000Z',
  },
]

describe('timeline url state', () => {
  for (const { description, params, expected } of testCases) {
    test(description, () => {
      const timelineState = newTimelineState(params, [])
      expect(getSearchParams(timelineState).toString()).toEqual(expected)
    })
  }
})

describe('history range', () => {
  const testCases = [
    {
      description: '1m',
      params: new NicerURLSearchParams({ duration: '1m' }),
})

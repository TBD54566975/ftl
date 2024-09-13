import { describe, expect, test } from "vitest";
import { NicerURLSearchParams, TimelineState } from "./timeline-state";
//
const testCases = [
  {
    description: "empty",
    params: new NicerURLSearchParams(),
    expected: ""
  },
  {
    description: "tailing (default state)",
    params: new NicerURLSearchParams({ tail: "1", paused: "0" }),
    expected: ""
  },
  {
    description: "modules",
    params: new NicerURLSearchParams({ modules: "time,echo" }),
    expected: "modules=time,echo"
  },
  {
    description: "log level",
    params: new NicerURLSearchParams({ log: "info" }),
    expected: "log=info"
  },
  {
    description: "time range",
    params: new NicerURLSearchParams({
      after: "2021-09-01T00:00:00.000Z",
      before: "2021-09-02T00:00:00.000Z",
    }),
    expected: "after=2021-09-01T00:00:00.000Z&before=2021-09-02T00:00:00.000Z"
  },
  {
    description: "tailing paused",
    params: new NicerURLSearchParams({ tail: "1", paused: "1" }),
    expected: "paused=1"
  },
  {
    description: "tailing with time range (incompatible settings--tail/paused should be ignored)",
    params: new NicerURLSearchParams({
      tail: "1",
      paused: "1",
      after: "2021-09-01T00:00:00.000Z",
      before: "2021-09-02T00:00:00.000Z",
    }),
    expected: "after=2021-09-01T00:00:00.000Z&before=2021-09-02T00:00:00.000Z"
  }
];

describe("timeline url state", () => {
  testCases.forEach(({ description, params, expected }) => {
    test(description, () => {
      const timelineState = new TimelineState(params, []);
      expect(timelineState.getSearchParams().toString()).toEqual(expected);
    });
  });
});

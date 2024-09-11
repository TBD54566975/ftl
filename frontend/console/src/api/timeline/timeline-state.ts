import { Timestamp } from "@bufbuild/protobuf";
import { EventsQuery_DeploymentFilter, EventsQuery_EventTypeFilter, EventsQuery_Filter, EventsQuery_LogLevelFilter, EventsQuery_TimeFilter, EventType, LogLevel } from "../../protos/xyz/block/ftl/v1/console/console_pb";
import { eventTypesFilter, logLevelFilter, modulesFilter, specificEventIdFilter, timeFilter } from "../../api/timeline";
import { TIME_RANGES, TimeRange } from "../../features/timeline/filters/TimelineTimeControls";

enum UrlKeys {
  ID = 'id',
  MODULES = 'modules',
  LOG = 'log',
  TYPES = 'types',
  PAUSED = 'paused',
  TAIL = 'tail',
  AFTER = 'after',
  BEFORE = 'before',
}

// spitballing the different time states
// desc       | tail   | paused | olderThan | newerThan
// -----------|--------|--------|-----------|----------
// last 5m    | false  | NA     | now - 5m  | now
// tail pause | true   | true   | NA        | NA
// tailing    | true   | false  | NA        | NA

/* states:
live paused
live tailing
range
*/

// type Range = {
//   olderThan?: Timestamp;
//   newerThan?: Timestamp;
// }

// TODO: type TimeState = Range | Tail;

// Hides the complexity of the URLSearchParams API and protobuf types.
export class TimelineState {
  isTailing = true;
  isPaused = false;
  timeRange: TimeRange = TIME_RANGES.tail;
  olderThan?: Timestamp;
  newerThan?: Timestamp;
  modules: string[] = [];
  logLevel?: LogLevel;
  eventTypes: EventType[] = [];
  eventId?: bigint;

  constructor(params: URLSearchParams) {
    // Quietly ignore invalid values from the user's URL...
    for (const [key, value] of params.entries()) {
      if (key === UrlKeys.ID) {
        this.eventId = BigInt(value);
      } else if (key === UrlKeys.MODULES) {
        this.modules = value.split(',');
      } else if (key === UrlKeys.LOG) {
        const enumValue = logValueToEnum(value)
        if (enumValue) {
          this.logLevel = enumValue;
        }
      } else if (key === UrlKeys.TYPES) {
        const types = value.split(',')
          .map((type) => eventTypeValueToEnum(type))
          .filter((type) => type !== undefined);
        if (types.length !== 0) {
          this.eventTypes = types
        }
      } else if (key === UrlKeys.PAUSED) {
        this.isPaused = value === '1';
      } else if (key === UrlKeys.TAIL) {
        this.isTailing = value === '1';
      } else if (key === UrlKeys.AFTER) {
        this.olderThan = Timestamp.fromDate(new Date(value));
      } else if (key === UrlKeys.BEFORE) {
        this.newerThan = Timestamp.fromDate(new Date(value));
      }
    }

    // TODO
    // this.timeRange = this.calculateTimeRange();

    // If we're loading a specific event, we don't want to tail.
    //     setSelectedTimeRange(TIME_RANGES['5m'])
    //     setIsTimelinePaused(true)
    //
  }

  getFilters(): EventsQuery_Filter[] {
    const filters: EventsQuery_Filter[] = [];
    if (this.eventId) {
      filters.push(specificEventIdFilter(this.eventId));
    }
    if (this.modules.length > 0) {
      filters.push(modulesFilter(this.modules));
    }
    if (this.logLevel) {
      filters.push(logLevelFilter(this.logLevel));
    }
    if (this.eventTypes.length > 0) {
      filters.push(eventTypesFilter(this.eventTypes));
    }
    if (this.olderThan || this.newerThan) {
      filters.push(timeFilter(this.olderThan, this.newerThan));
    }
    return filters;
  }

  getSearchParams() {
    const params = new NicerURLSearchParams();

    if (this.eventId) {
      params.set(UrlKeys.ID, this.eventId.toString());
    }
    if (this.modules.length > 0) {
      params.set(UrlKeys.MODULES, this.modules.join(','));
    }
    if (this.logLevel) {
      const logString = logEnumToValue(this.logLevel);
      if (logString) {
        params.set(UrlKeys.LOG, logString);
      }
    }
    if (this.eventTypes.length > 0) {
      const eventTypes = this.eventTypes
        .map((type) => eventTypeEnumToValue(type))
        .filter((type) => type !== undefined);
      if (eventTypes.length !== 0) {
        params.set(UrlKeys.TYPES, eventTypes.join(','));
      }
    }
    if (this.olderThan) {
      params.set(UrlKeys.AFTER, this.olderThan.toDate().toISOString());
    }
    if (this.newerThan) {
      params.set(UrlKeys.BEFORE, this.newerThan.toDate().toISOString());
    }
    if (this.isPaused) {
      params.set(UrlKeys.PAUSED, '1');
    }
    // // tailing is on by default, so we only need to set it if it's off.
    if (!this.isTailing) {
      params.set(UrlKeys.TAIL, '0');
    }

    return params;
  }

  // calculateTimeRange(): TimeRange {
  //   // Since we don't have the originally stated time range, we can make a guess based on the time range given.
  //   // If the user has modified the time range, we'll just use that as an injected "Custom (X minutes)" option in
  //   // the dropdown.
  //   if (this.time.olderThan && this.time.newerThan) {
  //     const ms = this.time.newerThan.toDate().getTime() - this.time.olderThan.toDate().getTime();

  //     return {
  //       label: 'Custom',
  //       value: ms,
  //     };
  //   }

  //   return TIME_RANGES.tail;
  // }
}

function logValueToEnum(value: string): LogLevel | undefined {
  switch (value) {
    case 'trace': return LogLevel.TRACE;
    case 'debug': return LogLevel.DEBUG;
    case 'info': return LogLevel.INFO;
    case 'warn': return LogLevel.WARN;
    case 'error': return LogLevel.ERROR;
    default: return undefined;
  }
}

function logEnumToValue(level: LogLevel): string | undefined {
  switch (level) {
    case LogLevel.TRACE: return 'trace';
    case LogLevel.DEBUG: return 'debug';
    case LogLevel.INFO: return 'info';
    case LogLevel.WARN: return 'warn';
    case LogLevel.ERROR: return 'error';
    default: return undefined
  }
}

function eventTypeValueToEnum(value: string): EventType | undefined {
  switch (value) {
    case 'log': return EventType.LOG;
    case 'call': return EventType.CALL;
    case 'created': return EventType.DEPLOYMENT_CREATED;
    case 'updated': return EventType.DEPLOYMENT_UPDATED;
    default: return undefined;
  }
}

function eventTypeEnumToValue(type: EventType): string | undefined {
  switch (type) {
    case EventType.LOG: return 'log';
    case EventType.CALL: return 'call';
    case EventType.DEPLOYMENT_CREATED: return 'created';
    case EventType.DEPLOYMENT_UPDATED: return 'updated';
    default: return undefined;
  }
}

export class NicerURLSearchParams extends URLSearchParams {
  toString(): string {
    // sort automatically for more predictable URLs
    this.sort();

    let s = super.toString();

    // we don't want to encode commas in the URL, so we replace '%2C' with ','
    s = s.replace(/%2C/g, ',');
    // similar with : and %3A in dates
    s = s.replace(/%3A/g, ':');

    return s;
  }
}

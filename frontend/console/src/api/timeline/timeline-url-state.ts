import { Timestamp } from "@bufbuild/protobuf";
import { EventsQuery_DeploymentFilter, EventsQuery_EventTypeFilter, EventsQuery_Filter, EventsQuery_LogLevelFilter, EventsQuery_TimeFilter, EventType, LogLevel } from "../../protos/xyz/block/ftl/v1/console/console_pb";
import { eventTypesFilter, logLevelFilter, modulesFilter, specificEventIdFilter } from "../../api/timeline";
import { TimeSettings } from "../../features/timeline/filters/TimelineTimeControls";

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

export class TimelineUrlState {
  time: TimeSettings = { isTailing: true, isPaused: false };
  filters: EventsQuery_Filter[] = [];

  constructor(params: URLSearchParams) {
    for (const [key, value] of params.entries()) {
      if (key === UrlKeys.ID) {
        this.filters.push(specificEventIdFilter(BigInt(value)));
      } else if (key === UrlKeys.MODULES) {
        this.filters.push(modulesFilter(value.split(',')));
      } else if (key === UrlKeys.LOG) {
        const enumValue = logValueToEnum(value)
        if (enumValue) {
          this.filters.push(logLevelFilter(enumValue));
        }
      } else if (key === UrlKeys.TYPES) {
        const types = value.split(',')
          .map((type) => eventTypeValueToEnum(type))
          .filter((type) => type !== undefined);
        if (types.length !== 0) {
          this.filters.push(eventTypesFilter(types as EventType[]));
        }
      } else if (key === UrlKeys.PAUSED) {
        this.time.isPaused = value === '1';
      } else if (key === UrlKeys.TAIL) {
        this.time.isTailing = value === '1';
      } else if (key === UrlKeys.AFTER) {
        this.time.olderThan = Timestamp.fromDate(new Date(value));
      } else if (key === UrlKeys.BEFORE) {
        this.time.newerThan = Timestamp.fromDate(new Date(value));
      }
    }

    console.log('todo: handle time range')
    // If we're loading a specific event, we don't want to tail.
    //     setSelectedTimeRange(TIME_RANGES['5m'])
    //     setIsTimelinePaused(true)
    //
  }

  getSearchParams() {
    const params = new NicerURLSearchParams();

    for (const filter of this.filters) {
      const ff = filter.filter;
      console.log('ff', ff)
      console.log('ff instanceof EventsQuery_LogLevelFilter', ff instanceof EventsQuery_LogLevelFilter)
      console.log('ff instanceof EventsQuery_DeploymentFilter', ff.value instanceof EventsQuery_DeploymentFilter)
      if (ff.value instanceof EventsQuery_LogLevelFilter) {
        const logLevel = logEnumToValue(ff.value.logLevel);
        if (logLevel) {
          params.set('log', logLevel);
        }
      } else if (ff.value instanceof EventsQuery_DeploymentFilter) {
        console.log('todo: handle deployment filter')
        const modules = ff.value.deployments.join(',');
        if (modules) {
          params.set('modules', modules);
        }
      } else if (ff.value instanceof EventsQuery_EventTypeFilter) {
        const eventTypes = ff.value.eventTypes
          .map((type) => eventTypeEnumToValue(type))
          .filter((type) => type !== undefined);
        if (eventTypes.length !== 0) {
          params.set('types', eventTypes.join(','));
        }
      } else if (ff.value instanceof EventsQuery_TimeFilter) {
        const newerThan = ff.value.newerThan;
        if (newerThan) {
          params.set('before', newerThan.toDate().toISOString());
        }

        const olderThan = ff.value.olderThan;
        if (olderThan) {
          params.set('after', olderThan.toDate().toISOString());
        }
      }
    }

    if (this.time.isPaused) {
      params.set('paused', '1');
    }
    // tailing is on by default, so we only need to set it if it's off.
    if (!this.time.isTailing) {
      params.set('tail', '0');
    }
    if (this.time.olderThan) {
      params.set('after', this.time.olderThan.toDate().toISOString());
    }
    if (this.time.newerThan) {
      params.set('before', this.time.newerThan.toDate().toISOString());
    }

    return params;
  }
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
    // sort automatically
    this.sort();

    const s = super.toString();

    // we don't want to encode commas in the URL, so we replace '%2C' with ','
    return s.replace(/%2C/g, ',');
  }
}

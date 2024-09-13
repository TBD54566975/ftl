import { useMemo } from 'react';
import { TimelineState } from '../../api/timeline/timeline-state';
import { useSearchParams } from 'react-router-dom';

export function useTimelineState() {
  const [searchParams, setSearchParams] = useSearchParams();
  const timelineState = useMemo(() => new TimelineState(searchParams, []), [searchParams]);

  return [
    timelineState,
    (timelineState: TimelineState) => {
      // TimelineState will be modified in place, which react won't detect,
      // but... searchParams will be updated, which will trigger a new TimelineState above.
      console.error("this is not used")
      setSearchParams(timelineState.getSearchParams());
    }
  ];
}

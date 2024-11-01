import datetime
from dataclasses import dataclass

from ftl import verb


@dataclass
class TimeRequest:
    pass


@dataclass
class TimeResponse:
    time: datetime.datetime


@verb
def time(req: TimeRequest) -> TimeResponse:
    return TimeResponse(time=datetime.datetime.now())

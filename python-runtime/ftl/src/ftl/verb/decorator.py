import functools
from typing import Any, Callable, Optional, TypeVar, Union

from .model import Verb

F = TypeVar("F", bound=Callable[..., Any])


def verb(
    func: Optional[F] = None, *, export: bool = False
) -> Union[F, Callable[[F], F]]:
    def actual_decorator(fn: F) -> F:
        return functools.update_wrapper(Verb(fn, export=export), fn)

    if func is not None:
        return actual_decorator(func)
    return actual_decorator

import functools
from typing import Any, Callable, Optional, TypeVar, Union

F = TypeVar("F", bound=Callable[..., Any])


def verb(func: Optional[F] = None) -> Union[F, Callable[[F], F]]:
    func._is_ftl_verb = True

    def actual_decorator(fn: F) -> F:
        # type_hints = get_type_hints(fn)
        # sig = inspect.signature(fn)
        # first_param = next(iter(sig.parameters))

        # self._verb_registry[fn.__name__] = {
        #     "func": fn,
        #     "input_type": type_hints[first_param],
        #     "output_type": type_hints["return"],
        # }

        @functools.wraps(fn)
        def wrapper(*args, **kwargs):
            return fn(*args, **kwargs)

        return wrapper

    if func is not None:
        return actual_decorator(func)

    return actual_decorator

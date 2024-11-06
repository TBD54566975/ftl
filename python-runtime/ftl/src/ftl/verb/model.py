import functools
import inspect
from typing import Any, Callable, Type, TypeVar, get_type_hints

F = TypeVar("F", bound=Callable[..., Any])


class Verb:
    def __init__(self, func: F, *, export: bool = False) -> None:
        self.func = func
        self.export = export

        self._type_hints = get_type_hints(func)
        self._signature = inspect.signature(func)
        self._first_param = next(iter(self._signature.parameters))

    def get_input_type(self) -> Type:
        """Get the input type (first parameter type) of the verb."""
        return self._type_hints[self._first_param]

    def get_output_type(self) -> Type:
        """Get the output type (return type) of the verb."""
        return self._type_hints["return"]

    def __call__(self, *args, **kwargs):
        return self.func(*args, **kwargs)

    def __get__(self, obj, objtype=None):
        if obj is None:
            return self
        return functools.partial(self.__call__, obj)

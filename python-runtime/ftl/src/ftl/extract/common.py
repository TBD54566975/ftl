from typing import Any, Dict, List, Optional, Type

from ftl.protos.xyz.block.ftl.v1.schema import schema_pb2 as schemapb

from .context import LocalExtractionContext


def extract_type(
    local_ctx: LocalExtractionContext, type_hint: Type[Any]
) -> Optional[schemapb.Type]:
    """Extracts type information from Python type hints and maps it to schema types."""
    if isinstance(type_hint, list):
        return extract_slice(local_ctx, type_hint)

    elif isinstance(type_hint, dict):
        return extract_map(local_ctx, type_hint)

    elif type_hint is Any:
        return schemapb.Type(any=schemapb.Any())

    elif isinstance(type_hint, type):
        if (
            type_hint is str
            or type_hint is int
            or type_hint is bool
            or type_hint is float
        ):
            return extract_basic_type(type_hint)

        if hasattr(type_hint, "__bases__"):
            return extract_class_type(local_ctx, type_hint)

        if callable(type_hint):
            return extract_function_type(local_ctx, type_hint)

    # Handle parametric types (e.g., List[int], Dict[str, int]) - Optional, uncomment if needed
    # elif hasattr(type_hint, "__origin__"):
    #     return extract_parametric_type(local_ctx, type_hint)

    # TODO: raise exception for unsupported types
    return None


def extract_slice(
    local_ctx: LocalExtractionContext, type_hint: List[Any]
) -> Optional[schemapb.Type]:
    if isinstance(type_hint, list) and type_hint:
        element_type = extract_type(local_ctx, type_hint[0])  # Assuming non-empty list
        if element_type:
            return schemapb.Type(array=schemapb.Array(element=element_type))
    return None


def extract_map(
    local_ctx: LocalExtractionContext, type_hint: Dict[Any, Any]
) -> Optional[schemapb.Type]:
    if isinstance(type_hint, dict):
        key_type = extract_type(local_ctx, list(type_hint.keys())[0])
        value_type = extract_type(local_ctx, list(type_hint.values())[0])
        if key_type and value_type:
            return schemapb.Type(map=schemapb.Map(key=key_type, value=value_type))
    return None


def extract_basic_type(type_hint: Type[Any]) -> Optional[schemapb.Type]:
    type_map = {
        str: schemapb.Type(string=schemapb.String()),
        int: schemapb.Type(int=schemapb.Int()),
        bool: schemapb.Type(bool=schemapb.Bool()),
        float: schemapb.Type(float=schemapb.Float()),
    }
    return type_map.get(type_hint, None)


# Uncomment and implement parametric types if needed
# def extract_parametric_type(local_ctx: LocalExtractionContext, type_hint: Type[Any]) -> Optional[schemapb.Type]:
#     if hasattr(type_hint, "__args__"):
#         base_type = extract_type(local_ctx, type_hint.__origin__)
#         param_types = [extract_type(local_ctx, arg) for arg in type_hint.__args__]
#         if isinstance(base_type, schemapb.Ref):
#             base_type.type_parameters.extend(param_types)
#             return base_type
#     return None


def extract_class_type(
    local_ctx: LocalExtractionContext, type_hint: Type[Any]
) -> Optional[schemapb.Type]:
    ref = schemapb.Ref(name=type_hint.__name__, module=type_hint.__module__)
    local_ctx.add_needs_extraction(ref)
    return schemapb.Type(ref=ref)


def extract_function_type(
    local_ctx: LocalExtractionContext, type_hint: Type[Any]
) -> Optional[schemapb.Type]:
    ref = schemapb.Ref(name=type_hint.__name__, module=type_hint.__module__)
    local_ctx.add_needs_extraction(ref)
    return schemapb.Type(ref=ref)

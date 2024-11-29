import ast
from typing import Optional

from ftl.extract import LocalExtractionContext, extract_type
from ftl.protos.xyz.block.ftl.schema.v1 import schema_pb2 as schemapb

from ftl.verb.model import Verb


class VerbExtractor(ast.NodeVisitor):
    def __init__(
        self, context: LocalExtractionContext, module_name: str, file_path: str
    ):
        self.context = context
        self.module_name = module_name
        self.file_path = file_path

    def load_function(self, func_name: str) -> Optional[Verb]:
        """Load a function from the module and return it if it exists."""
        try:
            module = self.context.load_python_module(self.module_name, self.file_path)
            return getattr(module, func_name, None)
        except ImportError as e:
            print(f"Error importing module {self.module_name}: {e}")
            return None

    def visit_FunctionDef(self, node: ast.FunctionDef) -> None:
        """Visit a function definition and extract schema if it's a verb."""
        func = self.load_function(node.name)
        if func is None or not isinstance(func, Verb):
            return

        try:
            verb = schemapb.Verb(
                pos=schemapb.Position(
                    filename=self.file_path, line=node.lineno, column=node.col_offset
                ),
                name=node.name,
                request=extract_type(self.context, func.get_input_type()),
                response=extract_type(self.context, func.get_output_type()),
                export=func.export,
            )
            self.context.add_verb(self.module_name, verb)
        except Exception as e:
            print(f"Error extracting Verb: {e}")

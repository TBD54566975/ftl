import ast
from typing import Any, Optional, Type

from ftl.extract.common import extract_type
from ftl.extract.context import LocalExtractionContext
from ftl.protos.xyz.block.ftl.v1.schema import schema_pb2 as schemapb

class TransitiveExtractor(ast.NodeVisitor):
    def __init__(self, context: LocalExtractionContext, module_name: str, file_path: str):
        self.context = context
        self.module_name = module_name
        self.file_path = file_path

    def load_function(self, func_name: str):
        try:
            module = self.context.load_python_module(self.module_name, self.file_path)
            func = getattr(module, func_name, None)
            if func is None:
                print(f"Function {func_name} not found in {self.module_name}")
                return None
            return func
        except ImportError as e:
            print(f"Error importing module {self.module_name}: {e}")
            return None

    @staticmethod
    def convert_ast_annotation_to_type_hint(
            annotation_node: ast.AST,
    ) -> Optional[Type[Any]]:
        """Converts an AST annotation node to a Python type hint."""
        if isinstance(annotation_node, ast.Name):
            # Handle built-in types like int, str, etc.
            type_name = annotation_node.id
            try:
                return eval(type_name)  # Convert to actual type like 'int', 'str', etc.
            except NameError:
                return None
        # Handle other cases like ast.Subscript, etc. (extend this for complex types)
        return None

    def visit_ClassDef(self, node):
        if self.context.must_extract(self.module_name, node.name):
            lineno = node.lineno
            col_offset = node.col_offset
            export = False
            for decorator in node.decorator_list:
                if isinstance(decorator, ast.Name) and decorator.id == "export":
                    export = True

            # Extract fields and their types
            fields = []
            for class_node in node.body:
                if isinstance(
                        class_node, ast.AnnAssign
                ):  # Annotated assignment (field)
                    field_name = (
                        class_node.target.id
                        if isinstance(class_node.target, ast.Name)
                        else None
                    )
                    if field_name and class_node.annotation:
                        type_hint = self.convert_ast_annotation_to_type_hint(
                            class_node.annotation
                        )
                        if type_hint:
                            field_type = extract_type(self.context, type_hint)
                            if field_type:
                                field_schema = schemapb.Field(
                                    name=field_name, type=field_type
                                )
                                fields.append(field_schema)
                        # TODO: else:
                        # surface error; require type hint for everything

            data = schemapb.Data(
                pos=schemapb.Position(
                    filename=self.file_path, line=lineno, column=col_offset
                ),
                name=node.name,
                fields=fields,
                export=export,
            )

            # Add to context or perform further processing
            self.context.add_data(self.module_name, data)
            self.context.remove_needs_extraction(self.module_name, data.name)

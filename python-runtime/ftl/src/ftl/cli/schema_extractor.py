import argparse
import ast
import concurrent.futures
import os
import sys
import tomllib
from contextlib import contextmanager

from ftl.extract import (
    GlobalExtractionContext,
    TransitiveExtractor,
)
from ftl.verb import (
    VerbExtractor,
)

# analyzers is a list of lists, where each sublist contains analyzers that can run in parallel
analyzers = [
    [VerbExtractor],
    [TransitiveExtractor],
]

@contextmanager
def set_analysis_mode(path: str):
    original_sys_path = sys.path.copy()
    sys.path.append(path)
    try:
        yield
    finally:
        sys.path = original_sys_path


def get_module_name(ftl_dir: str) -> str:
    ftl_toml_path = os.path.join(ftl_dir, "ftl.toml")

    if not os.path.isfile(ftl_toml_path):
        raise FileNotFoundError(f"ftl.toml file not found in the specified module directory: {ftl_dir}")

    with open(ftl_toml_path, "rb") as f:
        config = tomllib.load(f)
        module_name = config.get("module")
        if module_name:
            return module_name
        else:
            raise ValueError("module name not found in ftl.toml")


def analyze_directory(module_dir: str):
    """Analyze all Python files in the given module_dir in parallel."""
    global_ctx = GlobalExtractionContext()

    file_paths = []
    for dirpath, _, filenames in os.walk(module_dir):
        for filename in filenames:
            if filename.endswith(".py"):
                file_paths.append(os.path.join(dirpath, filename))

    for analyzer_batch in analyzers:
        with concurrent.futures.ProcessPoolExecutor() as executor:
            future_to_file = {
                executor.submit(
                    analyze_file, global_ctx, file_path, analyzer_batch
                ): file_path
                for file_path in file_paths
            }

            for future in concurrent.futures.as_completed(future_to_file):
                file_path = future_to_file[future]
                try:
                    future.result()  # raise any exception that occurred in the worker process
                except Exception as exc:
                    print(f"failed to extract schema from {file_path}: {exc};")

    output_dir = os.path.join(module_dir, ".ftl")
    os.makedirs(output_dir, exist_ok=True)  # Create .ftl directory if it doesn't exist
    output_file = os.path.join(output_dir, "schema.pb")

    serialized_schema = global_ctx.to_module_schema(get_module_name(module_dir)).SerializeToString()
    with open(output_file, "wb") as f:
        f.write(serialized_schema)


def analyze_file(global_ctx: GlobalExtractionContext, file_path: str, analyzer_batch):
    """Analyze a single Python file using multiple analyzers in parallel."""
    module_name = os.path.splitext(os.path.basename(file_path))[0]
    file_ast = ast.parse(open(file_path).read())
    local_ctx = global_ctx.init_local_context()

    with concurrent.futures.ThreadPoolExecutor() as executor:
        futures = [
            executor.submit(
                run_analyzer,
                analyzer_class,
                local_ctx,
                module_name,
                file_path,
                file_ast,
            )
            for analyzer_class in analyzer_batch
        ]

        for future in concurrent.futures.as_completed(futures):
            try:
                future.result()
            except Exception as exc:
                print(f"Analyzer generated an exception: {exc} in {file_path}")


def run_analyzer(analyzer_class, context, module_name, file_path, file_ast):
    analyzer = analyzer_class(context, module_name, file_path)
    analyzer.visit(file_ast)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "module_dir", type=str, help="The Python module directory to analyze."
    )
    args = parser.parse_args()

    dir = args.module_dir
    with set_analysis_mode(dir):
        analyze_directory(dir)

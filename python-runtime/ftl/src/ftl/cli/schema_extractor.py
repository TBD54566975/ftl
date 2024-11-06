import argparse
import ast
import concurrent.futures
import os
import sys
from contextlib import contextmanager

from ftl.extract import (
    GlobalExtractionContext,
    TransitiveExtractor,
)
from ftl.verb import (
    VerbExtractor,
)

# analyzers is now a list of lists, where each sublist contains analyzers that can run in parallel
analyzers = [
    [VerbExtractor],
    [TransitiveExtractor],
]


@contextmanager
def set_analysis_mode(path):
    original_sys_path = sys.path.copy()
    sys.path.append(path)
    try:
        yield
    finally:
        sys.path = original_sys_path


def analyze_directory(module_dir):
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
                # else:
                #     print(f"File {file_path} analyzed successfully.")

    for ref_key, decl in global_ctx.deserialize().items():
        print(f"Extracted Decl:\n{decl}")


def analyze_file(global_ctx: GlobalExtractionContext, file_path, analyzer_batch):
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

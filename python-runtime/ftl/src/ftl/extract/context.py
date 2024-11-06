import importlib.util
import multiprocessing
import threading

from ftl.protos.xyz.block.ftl.v1.schema import schema_pb2 as schemapb


class LocalExtractionContext:
    """Local context for a single Python file."""

    def __init__(self, needs_extraction, verbs, data):
        self.verbs = verbs
        self.data = data
        self.needs_extraction = needs_extraction
        self.module_cache = {}
        self.cache_lock = threading.Lock()

    def add_verb(self, module_name, verb):
        """Add a verb to the shared verbs map."""
        ref_key = RefKey(module=module_name, name=verb.name)
        self.verbs[ref_key] = verb.SerializeToString()

    def add_data(self, module_name, data):
        """Add a verb to the shared verbs map."""
        ref_key = RefKey(module=module_name, name=data.name)
        self.data[ref_key] = data.SerializeToString()

    def add_needs_extraction(self, ref: schemapb.Ref):
        ref_key = RefKey(module=ref.module, name=ref.name)
        # Only add the key if it doesn't exist in the dictionary, not if it's False
        if ref_key not in self.needs_extraction:
            self.needs_extraction[ref_key] = True

    def remove_needs_extraction(self, module, name):
        ref_key = RefKey(module=module, name=name)
        self.needs_extraction[ref_key] = False

    def must_extract(self, module, name):
        ref_key = RefKey(module=module, name=name)
        return ref_key in self.needs_extraction

    def load_python_module(self, module_name, file_path):
        """Load a Python module dynamically and cache it locally."""
        with self.cache_lock:
            if file_path in self.module_cache:
                return self.module_cache[file_path]

            spec = importlib.util.spec_from_file_location(module_name, file_path)
            module = importlib.util.module_from_spec(spec)
            spec.loader.exec_module(module)
            self.module_cache[file_path] = module
            return module


class GlobalExtractionContext:
    """Global context across multiple files in a package."""

    def __init__(self):
        manager = multiprocessing.Manager()
        self.needs_extraction = manager.dict()
        self.verbs = manager.dict()
        self.data = manager.dict()

    def deserialize(self):
        deserialized_decls = {}
        for ref_key, serialized_decl in self.verbs.items():
            decl = schemapb.Verb()
            decl.ParseFromString(serialized_decl)
            deserialized_decls[ref_key] = decl
        for ref_key, serialized_decl in self.data.items():
            decl = schemapb.Data()
            decl.ParseFromString(serialized_decl)
            deserialized_decls[ref_key] = decl
        return deserialized_decls

    def init_local_context(self) -> LocalExtractionContext:
        return LocalExtractionContext(self.needs_extraction, self.verbs, self.data)


class RefKey:
    def __init__(self, module, name):
        self.module = module
        self.name = name

    def __eq__(self, other):
        if isinstance(other, RefKey):
            return self.module == other.module and self.name == other.name
        return False

    def __hash__(self):
        return hash((self.module, self.name))

    def __repr__(self):
        return f"RefKey(module={self.module}, name={self.name})"

from importlib.util import module_from_spec, spec_from_file_location
from pathlib import Path
import sys

_nlp_service_dir = Path(__file__).resolve().parent / "nlp-service"
sys.path.insert(0, str(_nlp_service_dir))

_spec = spec_from_file_location("nlp_service_main", _nlp_service_dir / "main.py")
if _spec is None or _spec.loader is None:
    raise RuntimeError("Unable to load nlp-service/main.py")

_module = module_from_spec(_spec)
_spec.loader.exec_module(_module)

app = _module.app

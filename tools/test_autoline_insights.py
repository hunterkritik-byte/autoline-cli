import tempfile
import unittest
from pathlib import Path

from autoline_insights import analyze


class InsightsTest(unittest.TestCase):
    def test_detects_languages_and_signals_without_exposing_values(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "app.py").write_text("# TODO\napi_key='super-secret'\n", encoding="utf-8")
            (root / "main.rs").write_text("fn main() {}\n", encoding="utf-8")
            (root / "pyproject.toml").write_text("[project]\nname='demo'\n", encoding="utf-8")
            result = analyze(root)
            self.assertEqual(result["languages"]["Python"], 1)
            self.assertEqual(result["languages"]["Rust"], 1)
            self.assertEqual(result["manifests"]["Python"], 1)
            self.assertEqual(result["todo_fixme_markers"], 1)
            self.assertEqual(result["credential_pattern_signals"], 1)
            self.assertNotIn("super-secret", str(result))


if __name__ == "__main__":
    unittest.main()

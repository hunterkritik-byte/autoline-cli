#!/usr/bin/env python3
import json
import tempfile
import unittest
from pathlib import Path

from autoline_insights import analyze


class InsightsTest(unittest.TestCase):
    def test_inventory_and_heuristics(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / "main.py").write_text("# TODO\napi_key=\"dummy\"\n", encoding="utf-8")
            (root / "package.json").write_text("{}", encoding="utf-8")
            (root / ".git").mkdir()
            (root / ".git" / "ignored.py").write_text("TODO", encoding="utf-8")
            result = analyze(root)
            self.assertEqual(result["files"], 2)
            self.assertEqual(result["languages"]["Python"], 1)
            self.assertEqual(result["manifests"]["Node.js"], 1)
            self.assertEqual(result["todo_fixme_markers"], 1)
            self.assertEqual(result["credential_pattern_signals"], 1)

    def test_json_shape(self):
        with tempfile.TemporaryDirectory() as tmp:
            result = analyze(Path(tmp))
            encoded = json.dumps(result)
            self.assertIn("languages", encoded)
            self.assertIn("credential_pattern_signals", encoded)


if __name__ == "__main__":
    unittest.main()

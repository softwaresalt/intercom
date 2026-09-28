"""Tests for the single canonical Go masker (032.001-T / 032.005-T,
AC-2.1 / AC-2.4).

AC-2.1: the canonical Go masker is defined in exactly one place
(scripts/lib/gomask.py), and 032.001-T's decision is recorded alongside the
superseded divergence note in scripts/check-retired-architecture.sh.
032.005-T: check-write-path-precondition.sh deletes its local clone and
imports the shared module.

Run: python3 -m unittest discover -s scripts/lib/tests -v
"""

from __future__ import annotations

import re
import unittest
from pathlib import Path

LIB_DIR = Path(__file__).resolve().parents[1]
REPO_ROOT = LIB_DIR.parents[1]
SCRIPTS_DIR = REPO_ROOT / 'scripts'
RETIRED_WRAPPER = SCRIPTS_DIR / 'check-retired-architecture.sh'
WRITE_PATH = SCRIPTS_DIR / 'check-write-path-precondition.sh'


class CanonicalMaskerDecisionTest(unittest.TestCase):
    """032.001-T: decision record lives beside the superseded note."""

    def setUp(self):
        self.text = RETIRED_WRAPPER.read_text(encoding='utf-8')

    def test_decision_is_recorded(self):
        self.assertIn('CANONICAL GO MASKER DECISION (032.001-T)', self.text)
        self.assertIn('retired-architecture SUPERSET', self.text)

    def test_superseded_note_is_marked(self):
        self.assertIn('SUPERSEDED by 032.001-T', self.text)

    def test_expected_write_path_deltas_are_enumerated(self):
        self.assertIn('EXPECTED WRITE-PATH VERDICT DELTAS', self.text)


class SingleDefinitionTest(unittest.TestCase):
    """AC-2.1: exactly one definition of the Go masker under scripts/."""

    def test_masker_is_defined_exactly_once(self):
        pattern = re.compile(r'^\s*def mask_go_non_code\(', re.MULTILINE)
        definitions = []
        for path in sorted(SCRIPTS_DIR.rglob('*')):
            if not path.is_file() or path.suffix not in ('.py', '.sh'):
                continue
            if 'tests' in path.relative_to(SCRIPTS_DIR).parts:
                continue
            count = len(pattern.findall(path.read_text(encoding='utf-8')))
            definitions.extend([path.relative_to(REPO_ROOT).as_posix()] * count)
        self.assertEqual(definitions, ['scripts/lib/gomask.py'])


class WritePathUsesSharedMaskerTest(unittest.TestCase):
    """032.005-T: write-path gate imports the canonical masker."""

    def setUp(self):
        self.text = WRITE_PATH.read_text(encoding='utf-8')

    def test_local_clone_is_deleted(self):
        self.assertNotIn('def mask_go_non_code(', self.text)

    def test_shared_masker_is_imported(self):
        self.assertRegex(self.text, r'(?m)^from gomask import mask_go_non_code\b')
        self.assertIn("'scripts' / 'lib'", self.text)


if __name__ == '__main__':
    unittest.main()

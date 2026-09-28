"""Tests for the retired-arch wrapper reduction and the re-derived selection
pathspec pin (032.004-T, AC-2.5 / AC-2.6 / AC-2.7).

The pin must stay SOURCE-TEXT-ANCHORED: it reads the real source of
select_repo_paths() / should_scan_repo_path() via inspect.getsource(), so an
edit to the pathspec or prefix set is observable by an outside process. It
must never degrade into a module-constant self-comparison, and the former
disk-read of scripts/check-retired-architecture.sh must be retired.

Run: python3 -m unittest discover -s scripts/lib/tests -v
"""

from __future__ import annotations

import inspect
import re
import sys
import unittest
from pathlib import Path

LIB_DIR = Path(__file__).resolve().parents[1]
REPO_ROOT = LIB_DIR.parents[1]
if str(LIB_DIR) not in sys.path:
    sys.path.insert(0, str(LIB_DIR))

import retired_arch  # noqa: E402

WRAPPER = REPO_ROOT / 'scripts' / 'check-retired-architecture.sh'


def _pin_without_literals(path: str) -> bool:  # pragma: no cover - probe only
    return path.startswith('vendor/')


def _select_without_literals():  # pragma: no cover - probe only
    return []


class SelectionPathspecPinTest(unittest.TestCase):
    def test_real_functions_satisfy_the_pin(self):
        status = retired_arch.selection_pathspec_pin(
            retired_arch.select_repo_paths, retired_arch.should_scan_repo_path
        )
        self.assertEqual(status, {
            'select_found': True,
            'guard_found': True,
            'pathspec_ok': True,
            'prefix_ok': True,
        })

    def test_pin_defaults_to_the_real_functions(self):
        self.assertEqual(
            retired_arch.selection_pathspec_pin(),
            retired_arch.selection_pathspec_pin(
                retired_arch.select_repo_paths, retired_arch.should_scan_repo_path
            ),
        )

    def test_pin_is_source_anchored_not_constant(self):
        # A function whose SOURCE lacks the literals must fail the pin. A
        # module-constant self-comparison could never observe this.
        status = retired_arch.selection_pathspec_pin(_select_without_literals, _pin_without_literals)
        self.assertTrue(status['select_found'])
        self.assertTrue(status['guard_found'])
        self.assertFalse(status['pathspec_ok'])
        self.assertFalse(status['prefix_ok'])

    def test_pin_fails_closed_when_source_is_unavailable(self):
        status = retired_arch.selection_pathspec_pin(len, len)
        self.assertEqual(status, {
            'select_found': False,
            'guard_found': False,
            'pathspec_ok': False,
            'prefix_ok': False,
        })

    def test_disk_read_hack_is_retired(self):
        source = inspect.getsource(retired_arch)
        self.assertNotIn("'check-retired-architecture.sh'", source)
        self.assertNotIn('def extract_function_region(', source)


class WrapperTest(unittest.TestCase):
    def setUp(self):
        self.text = WRAPPER.read_text(encoding='utf-8')
        self.code_lines = [
            line for line in self.text.splitlines() if not line.lstrip().startswith('#')
        ]

    def test_wrapper_has_no_embedded_python_heredoc(self):
        heredocs = [line for line in self.code_lines if re.search(r'<<-?\s*[\'"]?\w+', line)]
        self.assertEqual(heredocs, [])
        self.assertNotIn('def ', '\n'.join(self.code_lines))

    def test_wrapper_dispatches_to_the_extracted_module(self):
        self.assertIn('scripts/lib/retired_arch.py', '\n'.join(self.code_lines))
        for flag in ('--self-test)', '--self-test-integrity)'):
            self.assertIn(flag, self.text)


if __name__ == '__main__':
    unittest.main()

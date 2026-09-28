"""Unit tests for scripts/lib/retired_arch.py (032.003-T, AC-2.8).

The engine must be importable as an ordinary module: neither sys.argv nor
Path.cwd() may be read at import time (the two former heredoc bindings
`mode = sys.argv[1]` and `root = Path.cwd()`), the canonical masker must be
the shared gomask implementation rather than a clone, and the repo root must
come from an explicit resolver rather than the importer's cwd.

Run: python3 -m unittest discover -s scripts/lib/tests -v
"""

from __future__ import annotations

import ast
import contextlib
import importlib
import io
import os
import sys
import tempfile
import unittest
from pathlib import Path

LIB_DIR = Path(__file__).resolve().parents[1]
REPO_ROOT = LIB_DIR.parents[1]
if str(LIB_DIR) not in sys.path:
    sys.path.insert(0, str(LIB_DIR))

import gomask  # noqa: E402

MODULE_PATH = LIB_DIR / 'retired_arch.py'


def _import_fresh(argv: list[str], cwd: Path):
    saved_argv, saved_cwd = sys.argv, Path.cwd()
    sys.modules.pop('retired_arch', None)
    try:
        sys.argv = argv
        os.chdir(cwd)
        return importlib.import_module('retired_arch')
    finally:
        sys.argv = saved_argv
        os.chdir(saved_cwd)


class ImportTimeBindingTest(unittest.TestCase):
    def test_import_with_empty_argv_and_foreign_cwd(self):
        with tempfile.TemporaryDirectory() as tmp:
            module = _import_fresh([], Path(tmp))
        self.assertFalse(hasattr(module, 'mode'), 'module-level `mode` binding must be removed')
        self.assertFalse(hasattr(module, 'root'), 'module-level `root` binding must be removed')

    def test_no_module_level_argv_or_cwd_reads(self):
        tree = ast.parse(MODULE_PATH.read_text(encoding='utf-8'))
        offenders: list[str] = []

        def is_main_guard(node: ast.AST) -> bool:
            # `if __name__ == '__main__':` never executes on import.
            return (
                isinstance(node, ast.If)
                and isinstance(node.test, ast.Compare)
                and isinstance(node.test.left, ast.Name)
                and node.test.left.id == '__name__'
                and any(isinstance(c, ast.Constant) and c.value == '__main__' for c in node.test.comparators)
            )

        def visit(node: ast.AST, in_function: bool) -> None:
            if not in_function and is_main_guard(node):
                return
            nested = in_function or isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef, ast.Lambda))
            if not in_function and isinstance(node, ast.Attribute):
                owner = node.value
                if isinstance(owner, ast.Name):
                    if owner.id == 'sys' and node.attr == 'argv':
                        offenders.append(f'sys.argv at line {node.lineno}')
                    if owner.id == 'Path' and node.attr == 'cwd':
                        offenders.append(f'Path.cwd at line {node.lineno}')
                    if owner.id == 'os' and node.attr == 'getcwd':
                        offenders.append(f'os.getcwd at line {node.lineno}')
            for child in ast.iter_child_nodes(node):
                visit(child, nested)

        visit(tree, False)
        self.assertEqual(offenders, [], 'import-time argv/cwd reads found')


class ModuleSurfaceTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.mod = _import_fresh(['x'], REPO_ROOT)

    def test_uses_shared_canonical_masker(self):
        self.assertIs(self.mod.mask_go_non_code, gomask.mask_go_non_code)
        self.assertIs(self.mod.struct_tag_re, gomask.struct_tag_re)

    def test_module_level_state_is_carried(self):
        self.assertEqual(
            set(self.mod.forbidden_parts),
            {'slack', 'socketmode', 'channel_id', 'team_id', 'acp', 'host_cli', 'ipc_name'},
        )
        self.assertEqual(
            [suite['name'] for suite in self.mod.fixture_suites],
            ['toml', 'go', 'go-differential'],
        )
        self.assertTrue(hasattr(self.mod, 'tomllib'))

    def test_resolve_repo_root_is_explicit(self):
        self.assertEqual(self.mod.resolve_repo_root(LIB_DIR).resolve(), REPO_ROOT.resolve())

    def test_resolve_repo_root_outside_a_repo_fails(self):
        with tempfile.TemporaryDirectory() as tmp:
            with self.assertRaises(SystemExit):
                self.mod.resolve_repo_root(Path(tmp))

    def test_unknown_mode_is_rejected(self):
        with self.assertRaises(SystemExit) as ctx:
            self.mod.main(['bogus-mode'], root=REPO_ROOT)
        self.assertIn('unknown mode', str(ctx.exception.code))

    def test_missing_mode_is_rejected(self):
        with self.assertRaises(SystemExit):
            self.mod.main([], root=REPO_ROOT)

    def test_integrity_mode_is_cwd_independent(self):
        buf = io.StringIO()
        saved = Path.cwd()
        with tempfile.TemporaryDirectory() as tmp:
            try:
                os.chdir(tmp)
                with contextlib.redirect_stdout(buf):
                    self.mod.main(['self-test-integrity'], root=REPO_ROOT)
            finally:
                os.chdir(saved)
        out = buf.getvalue()
        self.assertIn('PASS selection pathspec pin (AC-6/AG-1)', out)
        self.assertIn('PASS selection cmd/ coverage (AG-5/D4)', out)
        self.assertNotIn('FAIL', out)


if __name__ == '__main__':
    unittest.main()

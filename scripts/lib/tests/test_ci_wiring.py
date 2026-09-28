"""CI wiring and stdlib lint for the extracted gate engines (032.006-T,
AC-2.8).

ci.yml's lint job previously ran no Python lint/test step, so the extracted
modules' importability and these unit tests would never execute in CI. This
suite asserts the step exists, is blocking, runs on a pinned interpreter,
and is enumerated in the ci.yml LOCAL DIVERGENCE note. It also carries a
dependency-free lint pass (compile with warnings-as-errors plus an
unused-import check) so "lintable as an ordinary module" is executable
without installing a third-party linter.

Run: python3 -m unittest discover -s scripts/lib/tests -v
"""

from __future__ import annotations

import ast
import re
import unittest
import warnings
from pathlib import Path

LIB_DIR = Path(__file__).resolve().parents[1]
REPO_ROOT = LIB_DIR.parents[1]
CI_YML = REPO_ROOT / '.github' / 'workflows' / 'ci.yml'
TEST_CMD = 'python -m unittest discover -s scripts/lib/tests'


def _job_block(text: str, job_id: str) -> str:
    match = re.search(rf'(?m)^  {re.escape(job_id)}:\n(.*?)(?=^  [A-Za-z0-9_-]+:\n|\Z)', text, re.DOTALL)
    return match.group(1) if match else ''


def _step_block(job: str, command: str) -> str:
    steps = re.split(r'(?m)^      - ', job)
    for step in steps:
        if command in step:
            return step
    return ''


class CiWiringTest(unittest.TestCase):
    def setUp(self):
        self.text = CI_YML.read_text(encoding='utf-8')
        self.lint = _job_block(self.text, 'lint')

    def test_lint_job_runs_python_unit_tests(self):
        self.assertIn(TEST_CMD, self.lint)

    def test_python_step_is_blocking(self):
        step = _step_block(self.lint, TEST_CMD)
        self.assertTrue(step, 'python unit-test step not found in lint job')
        self.assertNotIn('continue-on-error', step)

    def test_lint_job_pins_python(self):
        self.assertRegex(self.lint, r'uses: actions/setup-python@[0-9a-f]{40}')
        self.assertRegex(self.lint, r"python-version: '3\.12'")
        # The pin only governs the unit-test step if it runs BEFORE it.
        self.assertLess(self.lint.index('actions/setup-python@'), self.lint.index(TEST_CMD))

    def test_step_is_enumerated_as_local_divergence(self):
        header = self.text.split('\nname: CI\n', 1)[0]
        self.assertIn('032.006-T', header)


class StdlibLintTest(unittest.TestCase):
    def python_files(self):
        files = sorted(LIB_DIR.glob('*.py')) + sorted((LIB_DIR / 'tests').glob('*.py'))
        self.assertTrue(files)
        return files

    def test_modules_compile_without_warnings(self):
        for path in self.python_files():
            with self.subTest(path=path.name), warnings.catch_warnings():
                warnings.simplefilter('error')
                compile(path.read_text(encoding='utf-8'), str(path), 'exec')

    def test_no_unused_imports(self):
        for path in self.python_files():
            with self.subTest(path=path.name):
                tree = ast.parse(path.read_text(encoding='utf-8'))
                imported: dict[str, int] = {}
                for node in ast.walk(tree):
                    if isinstance(node, ast.Import):
                        for alias in node.names:
                            imported[(alias.asname or alias.name).split('.')[0]] = node.lineno
                    elif isinstance(node, ast.ImportFrom) and node.module != '__future__':
                        for alias in node.names:
                            imported[alias.asname or alias.name] = node.lineno
                used = {n.id for n in ast.walk(tree) if isinstance(n, ast.Name)}
                exported: set[str] = set()
                for node in tree.body:
                    if isinstance(node, ast.Assign) and any(
                        isinstance(t, ast.Name) and t.id == '__all__' for t in node.targets
                    ):
                        exported = {elt.value for elt in node.value.elts if isinstance(elt, ast.Constant)}
                unused = sorted(name for name in imported if name not in used and name not in exported)
                self.assertEqual(unused, [])


if __name__ == '__main__':
    unittest.main()

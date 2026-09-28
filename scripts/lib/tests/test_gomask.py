"""Unit tests for scripts/lib/gomask.py (032.002-T, AC-2.1 / AC-2.8).

The canonical Go masker is the retired-architecture SUPERSET recorded by
032.001-T: raw_string buffering, struct_tag_re unmasking of a raw-string
literal whose ENTIRE content is a well-formed struct tag, and the
unterminated-raw-string EOF fail-closed branch.

Run: python3 -m unittest discover -s scripts/lib/tests -v
"""

from __future__ import annotations

import sys
import unittest
from pathlib import Path

LIB_DIR = Path(__file__).resolve().parents[1]
if str(LIB_DIR) not in sys.path:
    sys.path.insert(0, str(LIB_DIR))

import gomask  # noqa: E402


class MaskGoNonCodeTest(unittest.TestCase):
    def assert_shape_preserved(self, text: str, masked: str) -> None:
        # Masking preserves total length and line structure so finding line
        # numbers computed on the masked text match the source file.
        self.assertEqual(len(text), len(masked))
        self.assertEqual(text.count('\n'), masked.count('\n'))

    def test_code_passes_through_unchanged(self):
        text = 'package p\n\nfunc F() int { return 1 }\n'
        self.assertEqual(gomask.mask_go_non_code(text), text)

    def test_line_and_block_comments_are_masked(self):
        text = 'a := 1 // os.Remove channelID\n/* socketMode\nteamID */ b := 2\n'
        masked = gomask.mask_go_non_code(text)
        self.assert_shape_preserved(text, masked)
        for hidden in ('os.Remove', 'channelID', 'socketMode', 'teamID'):
            self.assertNotIn(hidden, masked)
        self.assertIn('a := 1', masked)
        self.assertIn('b := 2', masked)

    def test_interpreted_string_and_rune_are_masked(self):
        text = 's := "os.Create \\" channelID"\nr := \'\\\'\'\nx := 3\n'
        masked = gomask.mask_go_non_code(text)
        self.assert_shape_preserved(text, masked)
        self.assertNotIn('os.Create', masked)
        self.assertNotIn('channelID', masked)
        self.assertIn('x := 3', masked)

    def test_struct_tag_raw_string_is_unmasked(self):
        text = 'type T struct {\n\tA int `json:"channel_id" toml:"x"`\n}\n'
        masked = gomask.mask_go_non_code(text)
        self.assert_shape_preserved(text, masked)
        self.assertIn('json:"channel_id" toml:"x"', masked)
        self.assertNotIn('`', masked)

    def test_non_tag_raw_string_is_masked(self):
        text = 'q := `SELECT channel_id FROM t`\nz := 4\n'
        masked = gomask.mask_go_non_code(text)
        self.assert_shape_preserved(text, masked)
        self.assertNotIn('channel_id', masked)
        self.assertIn('z := 4', masked)

    def test_multiline_raw_string_is_masked(self):
        text = 'q := `json:"a"\njson:"b"`\n'
        masked = gomask.mask_go_non_code(text)
        self.assert_shape_preserved(text, masked)
        self.assertNotIn('json', masked)

    def test_unterminated_raw_string_at_eof_fails_closed(self):
        text = 'x := 1\ny := `json:"channel_id"'
        masked = gomask.mask_go_non_code(text)
        self.assert_shape_preserved(text, masked)
        self.assertNotIn('channel_id', masked)

    def test_struct_tag_grammar_is_strict(self):
        pattern = gomask.struct_tag_re
        self.assertTrue(pattern.match('json:"a"'))
        self.assertTrue(pattern.match(' json:"a"  toml:"b" '))
        self.assertFalse(pattern.match('1json:"a"'))
        self.assertFalse(pattern.match('json:"a"toml:"b"'))
        self.assertFalse(pattern.match('SELECT 1'))


if __name__ == '__main__':
    unittest.main()

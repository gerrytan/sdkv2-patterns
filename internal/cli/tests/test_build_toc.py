import tempfile
import unittest
from pathlib import Path

from svp.build_toc import build_toc


class BuildTocTests(unittest.TestCase):
    def setUp(self) -> None:
        self.temporary_directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary_directory.cleanup)
        self.root = Path(self.temporary_directory.name)
        self.readme = self.root / "README.md"
        self.readme.write_text(
            "# Patterns\n\n<!-- start-toc -->\nstale\n<!-- end-toc -->\n\nFooter\n",
            encoding="utf-8",
        )

    def write_markdown(self, name: str, content: str) -> None:
        path = self.root / name
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(content, encoding="utf-8")

    def test_groups_sorts_and_preserves_surrounding_content(self) -> None:
        self.write_markdown(
            "validation/z.md", "---\ntocTitle: Zebra\ntocSection: Validation\n---\n"
        )
        self.write_markdown(
            "validation/a b.md",
            '---\ntocTitle: "Alpha: [example]"\ntocSection: Validation\n---\n',
        )
        self.write_markdown(
            "basics.md", "---\ntocTitle: Intro\ntocSection: Basics\n---\n"
        )
        self.write_markdown("missing.md", "---\ntocTitle: Missing section\n---\n")
        self.write_markdown("plain.md", "# No metadata\n")
        self.write_markdown(
            ".venv/ignored.md", "---\ntocTitle: Ignored\ntocSection: Basics\n---\n"
        )

        build_toc(self.root)

        expected = (
            "# Patterns\n\n<!-- start-toc -->\n\n"
            "## Basics\n\n- [Intro](basics.md)\n\n"
            "## Validation\n\n"
            "- [Alpha: \\[example\\]](validation/a%20b.md)\n"
            "- [Zebra](validation/z.md)\n"
            "<!-- end-toc -->\n\nFooter\n"
        )
        self.assertEqual(self.readme.read_text(encoding="utf-8"), expected)
        build_toc(self.root)
        self.assertEqual(self.readme.read_text(encoding="utf-8"), expected)

    def test_empty_project_clears_stale_entries(self) -> None:
        build_toc(self.root)
        self.assertEqual(
            self.readme.read_text(encoding="utf-8"),
            "# Patterns\n\n<!-- start-toc -->\n\n<!-- end-toc -->\n\nFooter\n",
        )

    def test_invalid_markers_leave_readme_unchanged(self) -> None:
        for content in (
            "# No markers\n",
            "<!-- end-toc -->\n<!-- start-toc -->\n",
            "<!-- start-toc --><!-- start-toc --><!-- end-toc -->",
        ):
            with self.subTest(content=content):
                self.readme.write_text(content, encoding="utf-8")
                with self.assertRaises(ValueError):
                    build_toc(self.root)
                self.assertEqual(self.readme.read_text(encoding="utf-8"), content)

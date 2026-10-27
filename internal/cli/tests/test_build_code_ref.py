import tempfile
import unittest
from pathlib import Path

from svp.build_code_ref import build_code_ref


class BuildCodeRefTests(unittest.TestCase):
    def setUp(self) -> None:
        temporary_directory = tempfile.TemporaryDirectory()
        self.addCleanup(temporary_directory.cleanup)
        self.root = Path(temporary_directory.name)
        self.source = self.root / "source.go"
        self.source.write_text(
            "package example\n"
            "\t// ref-start: schema\n"
            '\t\t"sku": {\n'
            "\t\t\tRequired: true,\n"
            "\t\t},\n"
            "\t// ref-end: schema\n",
            encoding="utf-8",
        )
        self.markdown = self.root / "example.md"
        self.start = '<!-- ref-start ref="source.go#schema" lang="go" -->'

    def test_builds_multiple_references_and_is_repeatable(self) -> None:
        self.markdown.write_text(
            f"# Example\n\n{self.start}\nstale\n<!-- ref-end -->\n"
            f"\nBetween\n\n{self.start}\n<!-- ref-end -->\n\nFooter\n",
            encoding="utf-8",
        )
        expected_block = (
            f'{self.start}\n\n```go\n"sku": {{\n\tRequired: true,\n}},\n```\n\n<!-- ref-end -->'
        )
        expected = f"# Example\n\n{expected_block}\n\nBetween\n\n{expected_block}\n\nFooter\n"
        build_code_ref(self.root)
        self.assertEqual(self.markdown.read_text(encoding="utf-8"), expected)
        build_code_ref(self.root)
        self.assertEqual(self.markdown.read_text(encoding="utf-8"), expected)

    def test_includes_entire_file_and_is_repeatable(self) -> None:
        self.source.write_text(
            '    resource "hello_paper_plane" "test" {\n        name = "test"\n    }\n',
            encoding="utf-8",
        )
        start = '<!-- ref-start file="source.go" lang="terraform" -->'
        self.markdown.write_text(f"{start}\nstale\n<!-- ref-end -->\n", encoding="utf-8")
        expected = (
            f'{start}\n\n```terraform\nresource "hello_paper_plane" "test" {{\n'
            '    name = "test"\n}\n```\n\n<!-- ref-end -->\n'
        )
        build_code_ref(self.root)
        self.assertEqual(self.markdown.read_text(encoding="utf-8"), expected)
        build_code_ref(self.root)
        self.assertEqual(self.markdown.read_text(encoding="utf-8"), expected)

    def test_supports_space_indentation_and_other_languages(self) -> None:
        self.source.write_text(
            "    # ref-start: schema\n    if True:\n        print('hello')\n"
            "    # ref-end: schema\n",
            encoding="utf-8",
        )
        self.markdown.write_text(
            self.start.replace('lang="go"', 'lang="python"') + "\n<!-- ref-end -->",
            encoding="utf-8",
        )
        build_code_ref(self.root)
        self.assertIn(
            "```python\nif True:\n    print('hello')\n```",
            self.markdown.read_text(encoding="utf-8"),
        )

    def test_rejects_invalid_references_without_changing_markdown(self) -> None:
        for content in (
            self.start,
            "<!-- ref-end -->",
            f"{self.start}\n{self.start}\n<!-- ref-end -->",
            self.start.replace("#schema", "#missing") + "\n<!-- ref-end -->",
            self.start.replace("source.go", "missing.go") + "\n<!-- ref-end -->",
            self.start.replace(' lang="go"', "") + "\n<!-- ref-end -->",
            '<!-- ref-start file="missing.tf" lang="terraform" -->\n<!-- ref-end -->',
            '<!-- ref-start file="../outside.tf" lang="terraform" -->\n<!-- ref-end -->',
            self.start.replace(' lang="go"', ' file="source.go" lang="go"') + "\n<!-- ref-end -->",
        ):
            with self.subTest(content=content):
                self.markdown.write_text(content, encoding="utf-8")
                with self.assertRaises(ValueError):
                    build_code_ref(self.root)
                self.assertEqual(self.markdown.read_text(encoding="utf-8"), content)

    def test_skips_hidden_directories_and_unmarked_files(self) -> None:
        hidden = self.root / ".venv"
        hidden.mkdir()
        (hidden / "ignored.md").write_text(self.start, encoding="utf-8")
        self.markdown.write_text("# Unchanged\n", encoding="utf-8")
        build_code_ref(self.root)
        self.assertEqual(self.markdown.read_text(encoding="utf-8"), "# Unchanged\n")

from collections import defaultdict
from pathlib import Path
from urllib.parse import quote

import yaml


def build_toc(project_root: Path | None = None) -> None:
    """Populate the root README's TOC from Markdown YAML front matter."""
    if project_root is None:
        project_root = Path(__file__).resolve().parents[4]

    readme = project_root / "README.md"
    original = readme.read_text(encoding="utf-8")
    start_marker = "<!-- start-toc -->"
    end_marker = "<!-- end-toc -->"
    if original.count(start_marker) != 1 or original.count(end_marker) != 1:
        raise ValueError("README.md must contain exactly one pair of TOC markers")
    before, _, remaining = original.partition(start_marker)
    _, separator, after = remaining.partition(end_marker)
    if not separator:
        raise ValueError("README.md TOC markers are out of order")

    sections: dict[str, list[tuple[str, str]]] = defaultdict(list)
    for directory, directories, filenames in project_root.walk():
        directories[:] = [name for name in directories if not name.startswith(".")]
        for filename in filenames:
            path = directory / filename
            if path.suffix != ".md" or path == readme:
                continue
            lines = path.read_text(encoding="utf-8").splitlines()
            if not lines or lines[0] != "---":
                continue
            try:
                end = lines.index("---", 1)
            except ValueError:
                continue
            metadata = yaml.safe_load("\n".join(lines[1:end]))
            if not isinstance(metadata, dict):
                continue
            title = metadata.get("tocTitle")
            section = metadata.get("tocSection")
            if not all(isinstance(value, str) and value.strip() for value in (title, section)):
                continue
            relative_path = quote(path.relative_to(project_root).as_posix(), safe="/")
            sections[section.strip()].append((title.strip(), relative_path))

    toc = []
    for section in sorted(sections, key=lambda value: (value.casefold(), value)):
        toc.extend([f"## {section}", ""])
        for title, relative_path in sorted(
            sections[section], key=lambda entry: (entry[0].casefold(), entry)
        ):
            title = title.replace("\\", "\\\\").replace("[", "\\[").replace("]", "\\]")
            toc.append(f"- [{title}]({relative_path})")
        toc.append("")

    content = "\n".join(toc)
    updated = f"{before}{start_marker}\n\n{content}{end_marker}{after}"
    if updated != original:
        readme.write_text(updated, encoding="utf-8")

from pathlib import Path
import re
from textwrap import dedent


MARKER = re.compile(r"<!--\s*ref-(?P<kind>start|end)\b(?P<attributes>.*?)-->", re.DOTALL)
ATTRIBUTE = re.compile(r'''(\w+)\s*=\s*(["'])(.*?)\2''')


def _code_block(project_root: Path, attributes: str) -> str:
    properties = {match[1]: match[3] for match in ATTRIBUTE.finditer(attributes)}
    reference = properties.get("ref", "")
    language = properties.get("lang", "")
    source_name, separator, region = reference.rpartition("#")
    if not separator or not source_name or not region or not language:
        raise ValueError('ref-start requires ref="path#region" and lang="language"')
    if not re.fullmatch(r"[\w.+-]+", language):
        raise ValueError(f"Invalid code reference language: {language!r}")

    source = (project_root / source_name).resolve()
    if not source.is_relative_to(project_root.resolve()):
        raise ValueError(f"Code reference must be inside the project: {reference}")
    lines = source.read_text(encoding="utf-8").splitlines(keepends=True)
    # Support common line-comment prefixes in referenced source files.
    prefix = r"\s*(?://|#|--|/\*|\*|<!--)\s*"
    suffix = r"\s*(?:\*/|-->)?\s*"
    start = re.compile(prefix + r"ref-start:\s*" + re.escape(region) + suffix)
    end = re.compile(prefix + r"ref-end:\s*" + re.escape(region) + suffix)
    starts = [index for index, line in enumerate(lines) if start.fullmatch(line)]
    ends = [index for index, line in enumerate(lines) if end.fullmatch(line)]
    if len(starts) != 1 or len(ends) != 1 or starts[0] >= ends[0]:
        raise ValueError(f"Expected one ordered pair of source markers for {reference}")

    code = dedent("".join(lines[starts[0] + 1:ends[0]])).strip("\r\n")
    # Use a longer fence if the source itself contains Markdown fences.
    fence = "`" * max(3, max((len(run) + 1 for run in re.findall(r"`+", code)), default=0))
    return f"{fence}{language}\n{code}\n{fence}"


def build_code_ref(project_root: Path | None = None) -> None:
    """Populate Markdown reference markers with dedented source code blocks."""
    if project_root is None:
        project_root = Path(__file__).resolve().parents[4]

    updates = []
    for directory, directories, filenames in project_root.walk():
        directories[:] = [name for name in directories if not name.startswith(".")]
        for filename in sorted(filenames):
            path = directory / filename
            if path.suffix != ".md":
                continue
            original = path.read_text(encoding="utf-8")
            chunks = []
            previous_end = 0
            opening = None
            for marker in MARKER.finditer(original):
                if marker["kind"] == "start":
                    if opening is not None:
                        raise ValueError(f"Nested ref-start marker in {path}")
                    opening = marker
                else:
                    if opening is None:
                        raise ValueError(f"Unmatched ref-end marker in {path}")
                    try:
                        block = _code_block(project_root, opening["attributes"])
                    except (ValueError, OSError) as error:
                        raise ValueError(f"Invalid code reference in {path}: {error}") from error
                    chunks.extend([
                        original[previous_end:opening.end()],
                        f"\n\n{block}\n\n",
                        marker[0],
                    ])
                    previous_end = marker.end()
                    opening = None
            if opening is not None:
                raise ValueError(f"Unmatched ref-start marker in {path}")
            chunks.append(original[previous_end:])
            updated = "".join(chunks)
            if updated != original:
                updates.append((path, updated))

    for path, updated in updates:
        path.write_text(updated, encoding="utf-8")

import subprocess
from pathlib import Path


HOOK_HEADER = "#!/bin/sh\n# Installed by svp install-git-precommit.\n"

HOOK = HOOK_HEADER + """\
set -eu

cd "$(git rev-parse --show-toplevel)"
if ! git diff --cached --quiet --no-renames -- ':(glob)**/*.md' '.github/workflows/check-markdown.yml'; then
    echo 'Pre-commit: checking README table of contents...'
    uv run --locked --project internal/cli svp build-toc
    if ! git diff --exit-code -- README.md; then
        cat >&2 <<'EOF'
Error: README.md differs from the staged table of contents.
The TOC builder has run. Review README.md, then stage the updated TOC:
    git add README.md
Retry your commit after staging the update.
To regenerate the TOC manually:
    uv run --locked --project internal/cli svp build-toc
EOF
        exit 1
    fi
fi

if ! git diff --cached --quiet --no-renames -- ':(glob)**/*.md' ':(glob)internal/reference-provider/**' ':(glob)internal/cli/**' '.github/workflows/check-markdown.yml'; then
    echo 'Pre-commit: checking Markdown code references...'
    uv run --locked --project internal/cli svp build-code-ref
    if ! git diff --exit-code -- ':(glob)**/*.md'; then
        cat >&2 <<'EOF'
Error: Markdown code references differ from the staged content.
The code reference builder has run. Review the Markdown changes, then stage the updated files:
    git add <updated-markdown-files>
Retry your commit after staging the updates.
To regenerate code references manually:
    uv run --locked --project internal/cli svp build-code-ref
EOF
        exit 1
    fi
fi

if ! git diff --cached --quiet --no-renames -- ':(glob)internal/cli/**' '.github/workflows/test-cli.yml'; then
    echo 'Pre-commit: running CLI unit tests...'
    uv run --locked --project internal/cli python -m unittest discover -s internal/cli/tests -v
fi
"""


def install_git_precommit(project_root: Path | None = None) -> None:
    """Install the validation hook, respecting Git's configured hooks path."""
    if project_root is None:
        project_root = Path(__file__).resolve().parents[4]
    result = subprocess.run(
        ["git", "rev-parse", "--path-format=absolute", "--git-path", "hooks/pre-commit"],
        cwd=project_root,
        check=True,
        capture_output=True,
        text=True,
    )
    hook_path = Path(result.stdout.strip())
    if hook_path.exists():
        try:
            answer = input(f"Overwrite existing pre-commit hook at {hook_path}? [y/N] ")
        except EOFError:
            answer = ""
        if answer.strip().lower() not in {"y", "yes"}:
            print("Installation cancelled; existing pre-commit hook preserved.")
            return
    hook_path.parent.mkdir(parents=True, exist_ok=True)
    hook_path.write_text(HOOK, encoding="utf-8")
    hook_path.chmod(hook_path.stat().st_mode | 0o111)
    print(f"Installed pre-commit hook at {hook_path}")

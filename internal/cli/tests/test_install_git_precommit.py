import os
import subprocess
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

from sdkv2_cli.install_git_precommit import HOOK, HOOK_HEADER, install_git_precommit


class InstallGitPrecommitTests(unittest.TestCase):
    def setUp(self) -> None:
        temporary_directory = tempfile.TemporaryDirectory(prefix="cli hooks ")
        self.addCleanup(temporary_directory.cleanup)
        self.root = Path(temporary_directory.name)
        self.environment = {
            key: value for key, value in os.environ.items() if not key.startswith("GIT_")
        }
        environment_patch = patch.dict(os.environ, self.environment, clear=True)
        environment_patch.start()
        self.addCleanup(environment_patch.stop)
        self.git("init")
        self.hook_path = self.root / ".git/hooks/pre-commit"

    def git(self, *args: str) -> None:
        subprocess.run(
            ["git", *args], cwd=self.root, env=self.environment,
            check=True, capture_output=True, text=True,
        )

    def test_installs_executable_hook_and_can_be_repeated(self) -> None:
        install_git_precommit(self.root)
        with patch("builtins.input", return_value="y") as prompt:
            install_git_precommit(self.root)
        self.assertIn(str(self.hook_path), prompt.call_args.args[0])
        self.assertEqual(self.hook_path.read_text(encoding="utf-8"), HOOK)
        self.assertTrue(os.access(self.hook_path, os.X_OK))

    def test_respects_configured_hooks_path(self) -> None:
        self.git("config", "core.hooksPath", "custom hooks")
        install_git_precommit(self.root)
        hook = self.root / "custom hooks/pre-commit"
        self.assertEqual(hook.read_text(encoding="utf-8"), HOOK)
        self.assertFalse(self.hook_path.exists())

    def test_preserves_existing_custom_hook(self) -> None:
        self.hook_path.write_text("#!/bin/sh\nexit 0\n", encoding="utf-8")
        for answer in ("", "n", "no", "invalid"):
            with self.subTest(answer=answer), patch("builtins.input", return_value=answer):
                install_git_precommit(self.root)
                self.assertEqual(
                    self.hook_path.read_text(encoding="utf-8"), "#!/bin/sh\nexit 0\n"
                )

    def test_preserves_hook_when_input_is_unavailable(self) -> None:
        self.hook_path.write_text("#!/bin/sh\nexit 0\n", encoding="utf-8")
        with patch("builtins.input", side_effect=EOFError):
            install_git_precommit(self.root)
        self.assertEqual(
            self.hook_path.read_text(encoding="utf-8"), "#!/bin/sh\nexit 0\n"
        )

    def test_overwrites_custom_hook_when_confirmed(self) -> None:
        self.hook_path.write_text("#!/bin/sh\nexit 0\n", encoding="utf-8")
        with patch("builtins.input", return_value=" YES "):
            install_git_precommit(self.root)
        self.assertEqual(self.hook_path.read_text(encoding="utf-8"), HOOK)

    def test_updates_previously_installed_hook(self) -> None:
        self.hook_path.write_text(HOOK_HEADER + "exit 0\n", encoding="utf-8")
        with patch("builtins.input", return_value="yes"):
            install_git_precommit(self.root)
        self.assertEqual(self.hook_path.read_text(encoding="utf-8"), HOOK)

    def prepare_hook(self) -> tuple[Path, dict[str, str]]:
        install_git_precommit(self.root)
        bin_path = self.root / "bin"
        bin_path.mkdir()
        uv = bin_path / "uv"
        uv.write_text(
            '#!/bin/sh\n'
            'printf "%s\\n" "$*" >> "$UV_CALLS"\n'
            'case "$*" in\n'
            '  *buildToc)\n'
            '    if [ "${UPDATE_TOC:-0}" = 1 ]; then\n'
            '        printf "updated TOC\\n" > README.md\n'
            '    fi\n'
            '    exit "${BUILD_EXIT:-0}" ;;\n'
            '  *) exit "${TEST_EXIT:-0}" ;;\n'
            'esac\n',
            encoding="utf-8",
        )
        uv.chmod(0o755)
        calls = self.root / "uv-calls"
        environment = dict(
            self.environment,
            PATH=f"{bin_path}{os.pathsep}{os.environ['PATH']}",
            UV_CALLS=str(calls),
        )
        return calls, environment

    def test_hook_runs_validations_and_propagates_failures(self) -> None:
        calls, environment = self.prepare_hook()
        tracked = self.root / "README.md"
        tracked.write_text("original\n", encoding="utf-8")
        cli = self.root / "internal/cli/main.py"
        cli.parent.mkdir(parents=True)
        cli.write_text("# CLI\n", encoding="utf-8")
        self.git("add", "README.md")
        self.git("add", "internal/cli/main.py")

        for build_exit, test_exit, dirty, expected_exit, expected_calls in (
            (0, 0, False, 0, 2),
            (7, 0, False, 7, 1),
            (0, 9, False, 9, 2),
            (0, 0, True, 1, 1),
        ):
            with self.subTest(build_exit=build_exit, test_exit=test_exit, dirty=dirty):
                calls.write_text("", encoding="utf-8")
                tracked.write_text(
                    "changed\n" if dirty else "original\n", encoding="utf-8"
                )
                result = subprocess.run(
                    [str(self.hook_path)],
                    cwd=self.root,
                    env=dict(
                        environment, BUILD_EXIT=str(build_exit), TEST_EXIT=str(test_exit)
                    ),
                    capture_output=True,
                    text=True,
                )
                self.assertEqual(result.returncode, expected_exit, result.stderr)
                if dirty:
                    self.assertIn("git add README.md", result.stderr)
                    self.assertIn("Retry your commit", result.stderr)
                commands = calls.read_text(encoding="utf-8").splitlines()
                self.assertEqual(len(commands), expected_calls)
                self.assertEqual(
                    commands[0], "run --locked --project internal/cli sdkv2-cli buildToc"
                )
                if expected_calls == 2:
                    self.assertEqual(
                        commands[1],
                        "run --locked --project internal/cli python -m unittest discover -s internal/cli/tests -v",
                    )

    def test_runs_only_checks_for_relevant_staged_paths(self) -> None:
        calls, environment = self.prepare_hook()
        readme = self.root / "README.md"
        readme.write_text("original\n", encoding="utf-8")
        self.git("add", "README.md")
        self.git("-c", "user.name=Test", "-c", "user.email=test@example.com",
                 "-c", "core.hooksPath=/dev/null", "commit", "-m", "Initial")

        for filename, expected_command in (
            ("unrelated.txt", None),
            ("README.md", "buildToc"),
            ("validation/deep/example.md", "buildToc"),
            (".github/workflows/check-toc.yml", "buildToc"),
            ("internal/cli/src/example.py", "unittest"),
            ("internal/cli/uv.lock", "unittest"),
            (".github/workflows/test-cli.yml", "unittest"),
        ):
            with self.subTest(filename=filename):
                self.git("reset", "--quiet")
                readme.write_text("original\n", encoding="utf-8")
                calls.write_text("", encoding="utf-8")
                path = self.root / filename
                path.parent.mkdir(parents=True, exist_ok=True)
                path.write_text("changed\n", encoding="utf-8")
                self.git("add", filename)
                result = subprocess.run(
                    [str(self.hook_path)], cwd=self.root, env=environment,
                    capture_output=True, text=True,
                )
                self.assertEqual(result.returncode, 0, result.stderr)
                commands = calls.read_text(encoding="utf-8").splitlines()
                if expected_command is None:
                    self.assertEqual(commands, [])
                else:
                    self.assertEqual(len(commands), 1)
                    self.assertIn(expected_command, commands[0])

        # Unstaged relevant files must not trigger checks.
        self.git("reset", "--quiet")
        calls.write_text("", encoding="utf-8")
        result = subprocess.run(
            [str(self.hook_path)], cwd=self.root, env=environment,
            capture_output=True, text=True,
        )
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(calls.read_text(encoding="utf-8"), "")

    def test_generated_toc_must_be_staged_before_retrying(self) -> None:
        calls, environment = self.prepare_hook()
        readme = self.root / "README.md"
        readme.write_text("stale TOC\n", encoding="utf-8")
        self.git("add", "README.md")
        environment["UPDATE_TOC"] = "1"
        result = subprocess.run(
            [str(self.hook_path)], cwd=self.root, env=environment,
            capture_output=True, text=True,
        )
        self.assertEqual(result.returncode, 1)
        self.assertIn("git add README.md", result.stderr)
        self.assertIn("uv run --locked --project internal/cli sdkv2-cli buildToc", result.stderr)
        self.assertEqual(readme.read_text(encoding="utf-8"), "updated TOC\n")

        self.git("add", "README.md")
        result = subprocess.run(
            [str(self.hook_path)], cwd=self.root, env=environment,
            capture_output=True, text=True,
        )
        self.assertEqual(result.returncode, 0, result.stderr)

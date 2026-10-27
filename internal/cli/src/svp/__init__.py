import argparse

from .build_code_ref import build_code_ref
from .build_toc import build_toc
from .install_git_precommit import install_git_precommit


def main() -> None:
    parser = argparse.ArgumentParser(description="Internal SDK v2 patterns CLI.")
    commands = parser.add_subparsers(dest="command", required=True)
    build_toc_parser = commands.add_parser("build-toc", help="Build the table of contents in top README.md.")
    build_toc_parser.set_defaults(handler=build_toc)
    build_code_ref_parser = commands.add_parser(
        "build-code-ref", help="Build markdown code reference."
    )
    build_code_ref_parser.set_defaults(handler=build_code_ref)
    install_hook_parser = commands.add_parser(
        "install-git-precommit", help="Install the repository's pre-commit validation hook."
    )
    install_hook_parser.set_defaults(handler=install_git_precommit)
    args = parser.parse_args()
    args.handler()

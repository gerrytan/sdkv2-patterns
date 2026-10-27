import argparse

from .build_toc import build_toc


def main() -> None:
    parser = argparse.ArgumentParser(description="Internal SDK v2 patterns CLI.")
    commands = parser.add_subparsers(dest="command", required=True)
    build_toc_parser = commands.add_parser("buildToc", help="Build the table of contents.")
    build_toc_parser.set_defaults(handler=build_toc)
    args = parser.parse_args()
    args.handler()

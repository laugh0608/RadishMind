from __future__ import annotations

import subprocess
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

from scripts.checks.repository_governance import (
    find_markdown_link_errors,
    markdown_prose,
    repository_paths,
)


class RepositoryGovernanceTest(unittest.TestCase):
    @patch("scripts.checks.repository_governance.subprocess.run")
    def test_repository_paths_include_tracked_and_untracked_files(self, run_mock: unittest.mock.Mock) -> None:
        run_mock.return_value = subprocess.CompletedProcess(
            args=[],
            returncode=0,
            stdout=b"README.md\0docs/untracked.md\0",
        )

        paths = repository_paths(Path("/workspace"))

        self.assertEqual(paths, [Path("README.md"), Path("docs/untracked.md")])
        command = run_mock.call_args.args[0]
        self.assertIn("--cached", command)
        self.assertIn("--others", command)
        self.assertIn("--exclude-standard", command)

    def test_markdown_link_check_accepts_local_external_encoded_and_fenced_links(self) -> None:
        with tempfile.TemporaryDirectory() as temp_dir:
            repo_root = Path(temp_dir)
            docs_dir = repo_root / "docs"
            docs_dir.mkdir()
            (docs_dir / "guide.md").write_text("# 指南\n", encoding="utf-8")
            (docs_dir / "with space.md").write_text("# 空格路径\n", encoding="utf-8")
            (repo_root / "README.md").write_text(
                "\n".join(
                    (
                        "[指南](docs/guide.md#section)",
                        "[空格路径](<docs/with%20space.md>)",
                        "[外部链接](https://example.com/missing)",
                        "[页内锚点](#section)",
                        "`[行内代码](missing-inline.md)`",
                        "```markdown",
                        "[围栏示例](missing-fenced.md)",
                        "```",
                        "",
                    )
                ),
                encoding="utf-8",
            )

            errors = find_markdown_link_errors(repo_root, [Path("README.md")])

        self.assertEqual(errors, [])

    def test_markdown_link_check_rejects_broken_and_escaping_links(self) -> None:
        with tempfile.TemporaryDirectory() as temp_dir:
            repo_root = Path(temp_dir)
            (repo_root / "README.md").write_text(
                "[损坏链接](docs/missing.md)\n[越界链接](../outside.md)\n",
                encoding="utf-8",
            )

            errors = find_markdown_link_errors(repo_root, [Path("README.md")])

        self.assertEqual(
            errors,
            [
                "broken relative link: README.md -> docs/missing.md",
                "relative link escapes repository: README.md -> ../outside.md",
            ],
        )

    def test_markdown_prose_preserves_text_after_closed_fence(self) -> None:
        prose = markdown_prose("~~~text\n[忽略](missing.md)\n~~~\n[保留](README.md)\n")

        self.assertNotIn("missing.md", prose)
        self.assertIn("[保留](README.md)", prose)


if __name__ == "__main__":
    unittest.main()

from __future__ import annotations

import unittest

from scripts.checks.platform.check_platform_core_coverage import (
    is_postgres_specific_httpapi_file,
    parse_cover_profile,
)


class PlatformCoreCoverageTest(unittest.TestCase):
    def test_profile_uses_statement_counts_and_excludes_postgres_httpapi_files(self) -> None:
        summaries = parse_cover_profile(
            "\n".join(
                (
                    "mode: count",
                    "radishmind.local/services/platform/internal/httpapi/server.go:1.1,2.1 3 1",
                    "radishmind.local/services/platform/internal/httpapi/server.go:3.1,4.1 2 0",
                    "radishmind.local/services/platform/internal/httpapi/workflow_postgres_store.go:1.1,5.1 5 0",
                    "radishmind.local/services/platform/internal/bridge/bridge.go:1.1,3.1 4 2",
                )
            )
        )

        httpapi = summaries["internal/httpapi"]
        self.assertEqual(httpapi.covered_statements, 3)
        self.assertEqual(httpapi.total_statements, 5)
        self.assertEqual(httpapi.percentage, 60.0)
        self.assertEqual(httpapi.excluded_statements, 5)
        self.assertEqual(len(httpapi.excluded_files), 1)

        bridge = summaries["internal/bridge"]
        self.assertEqual(bridge.covered_statements, 4)
        self.assertEqual(bridge.total_statements, 4)
        self.assertEqual(bridge.excluded_statements, 0)

    def test_postgres_exclusion_is_scoped_by_package_and_filename_token(self) -> None:
        self.assertTrue(
            is_postgres_specific_httpapi_file(
                "internal/httpapi",
                "radishmind.local/services/platform/internal/httpapi/control_plane_postgres_repository.go",
            )
        )
        self.assertFalse(
            is_postgres_specific_httpapi_file(
                "internal/config",
                "radishmind.local/services/platform/internal/config/runtime_postgres_store.go",
            )
        )
        self.assertFalse(
            is_postgres_specific_httpapi_file(
                "internal/httpapi",
                "radishmind.local/services/platform/internal/httpapi/postgresql_compatibility.go",
            )
        )

    def test_invalid_profile_line_is_rejected(self) -> None:
        with self.assertRaisesRegex(ValueError, "invalid Go coverage profile line 2"):
            parse_cover_profile("mode: count\nnot-a-profile-line")


if __name__ == "__main__":
    unittest.main()

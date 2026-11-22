import json
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

from demo_release import assert_ready, check_render, prepare


class DemoReleaseTest(unittest.TestCase):
    sha = "a" * 40

    def test_stage_only_same_origin_proxy_and_ci_revision(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            prepare(root, self.sha, "https://demo-api.example")
            for app in ("admin", "frontend"):
                self.assertEqual((root / app / "public" / "revision.txt").read_text().strip(), self.sha)
            config = json.loads((root / "admin" / "vercel.json").read_text())
            self.assertEqual(config["rewrites"], [{
                "source": "/api/v1/:path*",
                "destination": "https://demo-api.example/api/v1/:path*",
            }])
            for origin in ("http://localhost", "https://name:secret@demo.example", "https://demo.example/path"):
                with self.assertRaises(ValueError):
                    prepare(root, self.sha, origin)

    def test_deployment_is_not_ready_until_live_and_on_exact_revision(self):
        self.assertFalse(check_render(self.sha, {"status": "build_in_progress"}))
        self.assertTrue(check_render(self.sha, {"status": "live", "commit": {"id": self.sha}}))
        for deployment in (
            {"status": "live"},
            {"status": "live", "commit": {"id": "b" * 40}},
            {"status": "build_failed", "commit": {"id": self.sha}},
        ):
            with self.assertRaises(RuntimeError):
                check_render(self.sha, deployment)

    def test_a_healthy_old_release_is_rejected(self):
        with patch("demo_release.request", return_value=(200, {"X-App-Revision": "b" * 40}, b'{"status":"ready"}')):
            with self.assertRaises(RuntimeError):
                assert_ready("https://demo-api.example", self.sha)
        with patch("demo_release.request", return_value=(200, {"X-App-Revision": self.sha}, b'{"status":"ready"}')):
            assert_ready("https://demo-api.example", self.sha)


if __name__ == "__main__":
    unittest.main()

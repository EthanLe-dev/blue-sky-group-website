"""Exercise command dispatch using a fake Docker CLI; never start real services."""
from pathlib import Path
import os
import subprocess
import tempfile
import unittest


@unittest.skipUnless(os.name == 'nt', 'Windows command launcher')
class ShortcutsTest(unittest.TestCase):
    def test_dispatch_and_failures(self):
        command = Path(__file__).resolve().with_name('bluesky.cmd')
        with tempfile.TemporaryDirectory(prefix='bluesky-command-') as directory:
            temp = Path(directory)
            log = temp / 'calls.txt'
            (temp / 'docker.cmd').write_text(
                '@echo off\necho %*>>"%BLUESKY_TEST_LOG%"\n'
                'if "%BLUESKY_TEST_FAIL%"=="1" exit /b 1\nexit /b 0\n'
            )
            env = {**os.environ, 'PATH': directory + os.pathsep + os.environ['PATH'],
                   'BLUESKY_TEST_LOG': str(log)}

            def invoke(action):
                return subprocess.run([os.environ['COMSPEC'], '/d', '/c', str(command), action],
                                      cwd=directory, env=env, capture_output=True, text=True)

            self.assertEqual(invoke('run').returncode, 0)
            self.assertIn('--profile app up -d --build', log.read_text())
            self.assertIn('--project-directory', log.read_text())
            log.unlink()
            self.assertEqual(invoke('off').returncode, 0)
            self.assertIn('--profile app --profile production-tools down', log.read_text())
            self.assertNotIn('-v', log.read_text())
            log.unlink()
            self.assertEqual(invoke('invalid').returncode, 2)
            self.assertFalse(log.exists())
            env['BLUESKY_TEST_FAIL'] = '1'
            self.assertEqual(invoke('run').returncode, 1)
            self.assertNotIn('compose', log.read_text())


if __name__ == '__main__':
    unittest.main()

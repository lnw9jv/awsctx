"""Exercise the real embedded picker in a PTY, keeping protocol stdout separate."""

import fcntl
import os
from pathlib import Path
import pty
import select
import struct
import subprocess
import sys
import tempfile
import termios
import time


def run_picker(binary, directory, defaults, key):
    master, slave = pty.openpty()
    fcntl.ioctl(slave, termios.TIOCSWINSZ, struct.pack("HHHH", 24, 80, 0, 0))
    env = dict(os.environ, HOME=str(directory), TERM="xterm-256color",
               AWS_CONFIG_FILE=str(directory / "config"),
               AWSCTX_STATE_DIR=str(directory / "state"), AWS_PROFILE="",
               FZF_DEFAULT_OPTS=defaults,
               FZF_DEFAULT_OPTS_FILE=str(directory / "fzf-options"))

    def controlling_terminal():
        os.setsid()
        fcntl.ioctl(0, termios.TIOCSCTTY, 0)

    process = subprocess.Popen([binary], env=env, stdin=slave,
                               stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                               preexec_fn=controlling_terminal)
    os.close(slave)
    screen = b""
    sent = False
    try:
        deadline = time.monotonic() + 10
        while process.poll() is None and time.monotonic() < deadline:
            if select.select([master], [], [], 0.1)[0]:
                try:
                    chunk = os.read(master, 65536)
                except OSError:
                    break
                screen += chunk
                if b"\x1b[6n" in chunk:
                    os.write(master, b"\x1b[1;1R")
                if not sent and b"dev" in screen and b"prod" in screen:
                    os.write(master, key)
                    sent = True
        out, err = process.communicate(timeout=2)
        assert sent, f"picker never displayed both profiles: stdout={out!r}, stderr={err!r}"
        if key == b"\r":
            assert process.returncode == 0, err
            assert out == b"export AWS_PROFILE=dev\n", out
            assert err == b"Switched to profile dev\n", err
        else:
            assert process.returncode != 0, "cancellation succeeded"
            assert out == b"", out
            assert b"cancelled" in err, err
    finally:
        if process.poll() is None:
            process.kill()
        process.communicate()
        os.close(master)


with tempfile.TemporaryDirectory(prefix="awsctx-picker-") as temporary:
    directory = Path(temporary)
    (directory / "config").write_text("[profile dev]\n[profile prod]\n")
    (directory / "fzf-options").write_text("--filter=prod\n")
    for defaults, key in [("--filter=prod", b"\r"), ("--multi --print-query", b"\r"),
                          ("", b"\x03")]:
        run_picker(sys.argv[1], directory, defaults, key)

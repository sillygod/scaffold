import argparse
import logging
import os
import subprocess
import sys
from abc import ABC, abstractmethod
from argparse import REMAINDER, ArgumentParser, Namespace
from threading import Thread
from typing import Any, Callable, Dict, List, NoReturn, Optional, Tuple, Type

from profiler.pusher import PProfPusher


def get_process_command(pid: int) -> str:
    try:
        # Use the 'ps' command to retrieve process information by PID
        output = subprocess.check_output(['ps', '-p', str(pid), '-o', 'command='])
        return output.decode('utf-8').strip()
    except subprocess.CalledProcessError:
        return ""


def get_child_processes(parent_pid):
    try:
        # Use the 'ps' command to retrieve process information by PID
        output = subprocess.check_output(['ps', '--ppid', str(parent_pid), '--no-headers', '-o', 'pid'])
        child_pids = [int(pid) for pid in output.decode('utf-8').strip().split('\n')]
        return child_pids
    except subprocess.CalledProcessError:
        return []


def pid_exists(pid):
    if pid < 0:
        return False  # NOTE: pid == 0 returns True
    try:
        os.kill(pid, 0)
    except ProcessLookupError:  # errno.ESRCH
        return False  # No such process
    except PermissionError:  # errno.EPERM
        return True  # Operation not permitted (i.e., process exists)
    else:
        return True  # no error, we can send a signal to the process


class AustinArgumentParser(ArgumentParser):
    """Austin Command Line parser.

    NOTE: this is copied from the package austin-python

    This command line parser is based on :class:`argparse.ArgumentParser` and
    provides a minimal implementation for parsing the standard Austin command
    line. The bool arguments of the constructor are used to specify whether
    the corresponding Austin option should be parsed or not. For example, if
    your application doesn't need the possiblity of switching to the
    alternative format, you can exclude this option with ``alt_format=False``.

    Note that al least one between ``pid`` and ``command`` is required, but
    they cannot be used together when invoking Austin.
    """

    def __init__(
        self,
        name: str = "austin",
        alt_format: bool = True,
        children: bool = True,
        exclude_empty: bool = True,
        exposure: bool = True,
        full: bool = True,
        interval: bool = True,
        memory: bool = True,
        pid: bool = True,
        sleepless: bool = True,
        timeout: bool = True,
        command: bool = True,
        **kwargs: Any,
    ) -> None:
        super().__init__(prog=name, **kwargs)

        def time(units: str) -> Callable[[str], int]:
            """Parse time argument with units."""
            base = int({"us": 1, "ms": 1e3, "s": 1e6}[units])

            def parser(arg: str) -> int:
                if arg.endswith("us"):
                    return int(arg[:-2]) // base
                if arg.endswith("ms"):
                    return int(arg[:-2]) * 1000 // base
                if arg.endswith("s"):
                    return int(arg[:-1]) * 1000000 // base
                return int(arg)

            return parser

        if not (pid and command):
            raise Exception("Austin command line parser must have at least one between pid " "and command.")

        if alt_format:
            self.add_argument(
                "-a",
                "--alt-format",
                help="Alternative collapsed stack sample format.",
                action="store_true",
            )

        if children:
            self.add_argument(
                "-C",
                "--children",
                help="Attach to child processes.",
                action="store_true",
            )

        if exclude_empty:
            self.add_argument(
                "-e",
                "--exclude-empty",
                help="Do not output samples of threads with no frame stacks.",
                action="store_true",
            )

        if exposure:
            self.add_argument(
                "-x",
                "--exposure",
                help="Sample for the given number of seconds only.",
                type=time("s"),
                default=None,
            )

        if full:
            self.add_argument(
                "-f",
                "--full",
                help="Produce the full set of metrics (time +mem -mem).",
                action="store_true",
            )

        if interval:
            self.add_argument(
                "-i",
                "--interval",
                help="Sampling interval (default is 100 μs).",
                type=time("us"),
            )

        if memory:
            self.add_argument("-m", "--memory", help="Profile memory usage.", action="store_true")

        if pid:
            self.add_argument(
                "-p",
                "--pid",
                help="The the ID of the process to which Austin should attach.",
                type=int,
            )

        if sleepless:
            self.add_argument("-s", "--sleepless", help="Suppress idle samples.", action="store_true")

        if timeout:
            self.add_argument(
                "-t",
                "--timeout",
                help="Approximate start up wait time. Increase on slow machines " "(default is 100 ms).",
                type=time("ms"),
            )

        if command:
            self.add_argument(
                "command",
                type=str,
                nargs=REMAINDER,
                help="The command to execute if no PID is provided, followed by " "its arguments.",
            )

    def parse_args(self, args: List[str], namespace: Optional[Namespace] = None) -> Namespace:  # type: ignore[override]
        """Parse the list of arguments.

        Return a :class:`argparse.Namespace` with the parsed result. If no PID
        nor a command are passed, an instance of the
        :class:`AustinCommandLineError` exception is thrown.
        """
        parsed_austin_args, unparsed = super().parse_known_args(args, namespace)
        if unparsed:
            raise Exception(f"Some arguments were left unparsed: {unparsed}")

        if not parsed_austin_args.pid and not parsed_austin_args.command:  # type: ignore
            raise Exception("No PID or command given.")

        return parsed_austin_args

    def exit(self, status: int = 0, message: str = "") -> NoReturn:
        """Raise exception on error."""
        raise Exception(message, status)

    @staticmethod
    def to_list(args: Namespace) -> List[str]:
        """Convert a :class:`argparse.Namespace` to a list of arguments.

        This is the opposite of the parsing of the command line. This static
        method is intended to filter and reconstruct the command line arguments
        that need to be passed to lower level APIs to start the actual Austin
        process.
        """
        arg_list = []
        if getattr(args, "alt_format", None):
            arg_list.append("-a")
        if getattr(args, "children", None):
            arg_list.append("-C")
        if getattr(args, "exclude_empty", None):
            arg_list.append("-e")
        if getattr(args, "full", None):
            arg_list.append("-f")
        if getattr(args, "interval", None):
            arg_list += ["-i", str(args.interval)]
        if getattr(args, "memory", None):
            arg_list.append("-m")
        if getattr(args, "pid", None):
            arg_list += ["-p", str(args.pid)]
        if getattr(args, "sleepless", None):
            arg_list.append("-s")
        if getattr(args, "timeout", None):
            arg_list += ["-t", str(args.timeout)]
        if getattr(args, "command", None):
            arg_list += args.command

        return arg_list


class AustinError(Exception):
    """Basic Austin Error."""

    pass


class AustinTerminated(AustinError):
    """Austin termination exception.

    Thrown when Austin is terminated with a call to ``terminate``.
    """

    pass


class AbstractExporter(ABC):
    """
    Defines the general API that abstract Austin as an external process.
    Subclasses should implement the :func:`start` method and either define the
    :func:`on_sample_received` method or pass it via the constructor.
    Additionally, the :func:`on_ready` and the :func:`on_terminate` methods can
    be overridden or passed via the constructor to catch the corresponding
    events. Austin is considered to be ready when the first sample is received;
    it is considered to have terminated when the process has terminated
    gracefully.

    If an error was encountered, the :class:`AustinError` exception is thrown.
    """

    def __init__(
        self,
        sample_callback: Optional[Callable[[bytes], None]] = None,
        ready_callback: Optional[Callable[[int, int, str], None]] = None,
        terminate_callback: Optional[Callable[[Dict[str, str]], None]] = None,
    ) -> None:
        self._running: bool = False
        self._meta: Dict[str, str] = {}

        self._sample_callback = sample_callback or self.on_sample_received  # type: ignore[attr-defined]
        self._terminate_callback = terminate_callback or self.on_terminate
        self._ready_callback = ready_callback or self.on_ready

    def _get_process_info(self, args: argparse.Namespace, austin_pid: int) -> Tuple[int, int, str]:
        if not pid_exists(austin_pid):
            raise AustinError("Cannot find Austin process.") from None

        if not pid_exists(args.pid):
            raise AustinError(f"Cannot attach to process with invalid PID {args.pid}.") from None

        pid = args.pid
        cmd_line = get_process_command(pid)
        self._running = True

        return austin_pid, pid, cmd_line

    @abstractmethod
    def start(self, args: Optional[List[str]] = None) -> Any:
        """Start Austin.

        Every subclass should implement this method and ensure that it spawns
        a new Austin process.
        """
        raise NotImplementedError

    def is_running(self) -> bool:
        """Determine whether Austin is running."""
        return self._running

    def submit_sample(self, data: bytes) -> None:
        """Submit a sample to the sample callback.

        This method takes care of converting the raw binary data retrieved from
        Austin into a Python string.
        """
        self._sample_callback(data)

    def check_exit(self, rcode: int, stderr: Optional[str]) -> None:
        """Check Austin exit status."""
        if rcode:
            if rcode in {-15, 15, 241}:
                raise AustinTerminated()
            raise AustinError(f"({rcode}) {stderr}")

    # ---- Default callbacks ----

    def on_terminate(self, stats: Dict[str, str]) -> Any:
        """Terminate event callback.

        Implement to be notified when Austin has terminated gracefully. The
        callback accepts an argument that will receive the global statistics.
        """
        pass

    def on_ready(
        self,
        process: int,
        child_process: int,
        command_line: str,
        data: Any = None,
    ) -> Any:
        """Ready event callback.

        Implement to get notified when Austin has successfully started or
        attached the Python process to profile and the first sample has been
        produced. This callback receives the Austin process and it's (main)
        profiled process as instances of :class:`psutil.Process`, along with
        the command line of the latter.
        """
        pass


class SimpleExporter(AbstractExporter):
    """Simple implementation of exporter.

    Example::

        class SongExporter(SimpleExporter):
            def on_ready(self, process, child_process, command_line):
                print(f"Austin PID: {process.pid}")
                print(f"Python PID: {child_process.pid}")
                print(f"Command Line: {command_line}")

            def on_sample_received(self, line):
                print(line)

            def on_terminate(self, data):
                print(data)

        try:
            s = SongExporter()
            s.start(["-i", "10000"], ["python3", "myscript.py"])
        except KeyboardInterrupt:
            pass
    """

    def _read_meta(self) -> Dict[str, str]:
        assert self.proc.stdout

        meta = {}

        while True:
            line = self.proc.stdout.readline().decode().strip()
            if not (line and line.startswith("# ")):
                break
            key, _, value = line[2:].partition(": ")
            meta[key] = value

        self._meta.update(meta)

        return meta

    def start(self, args: List[str], sudo: bool = False) -> None:
        while True:
            try:
                self._start(args, sudo=sudo)
            except Exception as e:
                logging.error(e)

    def _start(self, args: List[str], sudo: bool = False) -> None:
        """Start the Austin process."""
        program = ["sudo", "austin"] if sudo else ["austin"]
        try:
            self.proc = subprocess.Popen(
                program + (args or sys.argv[1:]),
                bufsize=-1,  # NOTE: check the impact of performance
                stdin=subprocess.PIPE,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
            )
        except FileNotFoundError:
            raise AustinError("Austin executable not found.") from None


        if not self.proc.stdout:
            raise AustinError("Standard output stream is unexpectedly missing")
        if not self.proc.stderr:
            raise AustinError("Standard error stream is unexpectedly missing")

        try:

            if not self._read_meta():
                raise AustinError("Austin did not start properly")

            self._ready_callback(*self._get_process_info(AustinArgumentParser().parse_args(args), self.proc.pid))

            while self.is_running():
                data = self.proc.stdout.readline().strip()
                if not data:
                    break

                self.submit_sample(data)


            self._terminate_callback(self._read_meta())
            try:
                stderr = self.proc.communicate(timeout=1)[1].decode().rstrip()
            except subprocess.TimeoutExpired:
                stderr = ""
            self.check_exit(self.proc.wait(), stderr)

        except Exception:
            self.proc.terminate()
            self.proc.wait()
            raise

        finally:
            self._running = False


class PyroscopeExporter(SimpleExporter):

    """PyroscopeExporter is an exporter sending the memory profile
    to the pyroscope server.
    """

    def __init__(self, profiler_uploader: PProfPusher, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self._pprof_pusher = profiler_uploader

    def on_ready(self, process, child_process, command_line):
        logging.info(f"austin pid: {process}, app pid: {child_process} cmd: {command_line}")
        # set the meta
        self._pprof_pusher.set_metric_type(self._meta["mode"])

    def on_sample_received(self, line: bytes):
        try:
            sample = line.decode().strip()
        except UnicodeDecodeError:
            try:
                sample = line.decode("ascii").strip()
            except UnicodeDecodeError:
                return

        logging.debug(f"pprof metric samples: {sample}")
        self._pprof_pusher.collect(sample)

    def on_terminate(self, data):
        logging.debug(f"PyroscopeExporter is terminating... {data}")


def spawn_thread_exporter(exporter: Type[SimpleExporter], args: List[str]):
    """ThreadExporter is a simple wrapper to make exporter not to
    block the main thread.

    """

    def _inner(*args: str) -> None:
        try:
            exp = exporter()
            exp.start(list(args))
        except KeyboardInterrupt:
            pass
        except Exception as e:
            raise e

    t = Thread(target=_inner, daemon=True, args=args)
    t.start()

import logging
import signal
import threading
import traceback
from threading import Timer
from typing import Callable, Optional

logger = logging.getLogger(__name__)


class ThreadTimer:

    """A simle thread spawner with specified time interval"""

    def __init__(self, seconds: float, target, *args, times=-1, **kwargs):
        self.seconds = seconds
        self.target = target
        self.kwargs = kwargs
        self.args = args
        self._limit_count = times  # -1 means infinite
        self._count = 0
        self.thread: Optional[Timer] = None

    def handler(self):
        try:
            self.target(*self.args, **self.kwargs)
        except Exception:
            logger.error(traceback.format_exc(chain=False))

        self.thread = Timer(self.seconds, self.handler)
        if self._limit_count > 0 and self._limit_count - 1 <= self._count:
            self.cancel()
            return

        if self._limit_count > 0:
            self._count += 1

        self.thread.start()

    def start(self):
        self.thread = Timer(self.seconds, self.handler)
        self.thread.start()

    def cancel(self):
        if self.thread is not None:
            self.thread.cancel()


class Threadmanager:

    """Threadmanager is a simple single process multithread manager"""

    def __init__(self):
        self._threads = []
        self._stop = threading.Event()
        self._stop_callback = []
        self._graceful_period = 30

    def add_stop_callback(self, cb: Callable):
        self._stop_callback.append(cb)

    def add_periodical_thread(self, procedure: Callable, *args, time: float = 1.0, times=-1, **kwargs):
        t = ThreadTimer(time, procedure, *args, times=times, **kwargs)
        t.start()

        def _inner():
            t.cancel()

        self.add_stop_callback(_inner)

    def add_thread(self, procedure: Callable, *args, daemon=False, **kwargs):
        t = threading.Thread(target=procedure, args=args, daemon=daemon, kwargs=kwargs)
        self._threads.append(t)

    def shutdown(self):
        logger.info("process the remaining data and release resources")

    def stop(self, signal, frame):
        logger.info(f"receiving signal: {signal}, {frame}, ready to graceful stop ")

        def _inner():
            self._stop.set()

        self._stop_timer = ThreadTimer(self._graceful_period, _inner, times=1)
        self._stop_timer.start()

        for cb in self._stop_callback:
            cb()

        # joins the non-daemon threads
        for th in self._threads:
            logger.info(f"wait the thread: {th}, daemon:{th.daemon}")
            if not th.daemon:
                th.join()

        # all resources are released so we can start to shutdown the manager
        self._stop.set()
        self._stop_timer.cancel()

    def start(self):
        for th in self._threads:
            th.start()

        self._stop.wait()

    def run(self, graceful_period: int = 50):
        self._graceful_period = graceful_period
        signal.signal(signal.SIGINT, self.stop)
        signal.signal(signal.SIGTERM, self.stop)

        try:
            self.start()
        except Exception:
            logger.error(traceback.format_exc())

        self.shutdown()

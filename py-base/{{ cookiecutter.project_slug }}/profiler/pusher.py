import io
import logging
import threading
from datetime import datetime
from typing import IO, List, Optional
from urllib.parse import urlencode

import requests
from app.utils import retry

from profiler.producer import Producer, MetricType, Sample


class RWLock:
    def __init__(self):
        self._lock = threading.Condition(threading.Lock())
        self._readers = 0

    def rlock(self):
        self._lock.acquire()
        try:
            self._readers += 1
        finally:
            self._lock.release()

    def runlock(self):
        self._lock.acquire()
        try:
            self._readers -= 1
            if not self._readers:
                self._lock.notifyAll()
        finally:
            self._lock.release()

    def lock(self):
        self._lock.acquire()
        while self._readers > 0:
            self._lock.wait()

    def unlock(self):
        self._lock.release()


class PProfPusher:

    """PProfPusher collects the samples and transformed them to pprof format
    and provide a function that push the data to pyropscope server via API
    post.

    currently, only can be used in single thread. for performance, this is
    designed to mitigate the lock usages.
    """

    def __init__(self, name, server_url: str, spy_name: str = 'austin', unit='samples'):
        self._uri = f"{server_url}/ingest?"
        self._rwlock = RWLock()
        self._metric_type: Optional[MetricType] = None
        self._buf: IO[str] = io.StringIO()
        self._params = {
            'name': name,  # application name
            # 'format': 'pprof',
            'from': None,
            'until': None,
            # 'sampleRate': 100,
            # 'aggregationType': 'sum',
            'spyName': spy_name,
            # 'units': 'samples',  # TODO: create an enum for this param
        }

    def set_metric_type(self, mode: str):
        self._mode = mode
        self._metric_type = MetricType.from_mode(mode)

    def generate_samples(self) -> List:
        samples = []
        for line in self._buf.readlines():
            try:
                samples.append(Sample.parse(line, self._metric_type))
            except Exception as e:
                logging.warning(f"unexpcted metric format: {line}, error: {e}")

        return samples

    def gen_uri(self) -> str:
        # api schema: https://grafana.com/docs/pyroscope/latest/configure-server/about-server-api/
        # s
        return f'{self._uri}{urlencode(self._params)}'

    def collect(self, sample: str):
        self._rwlock.lock()
        if self._params['from'] is None:
            self._params['from'] = int(datetime.now().timestamp())

        assert isinstance(self._buf, io.StringIO)
        self._buf.write(sample)
        self._buf.write('\n')
        self._rwlock.unlock()

    def push(self):
        """currently, intended to be designed for manually triggering by other object."""
        logging.info(f"before push: the content size: {self._buf.tell()}")
        if self._buf.tell() == 0:
            return

        # create a python request multipart/form-data
        self._rwlock.rlock()

        # we need to move the file read pointer to the head
        self._buf.seek(0)

        # transform the data to pprof format with the producer
        pprof = Producer(self._mode)

        pprof.create_profile_from_samples(self.generate_samples(), 1, 1)


        form_data = {
            "profile": ("profile.pprof", pprof.outputio().getvalue(), "application/octet-stream"),
        }

        self._params['until'] = int(datetime.now().timestamp())

        res, exc = retry(requests.post, self.gen_uri(), files=form_data)
        if exc:
            logging.error("exception when uploading profile: %s", exc)
        # res = requests.post(self.gen_uri(), files=form_data)

        if res and not res.ok:
            # NOTE: a conveninet testing website
            # requests.post("https://httpbin.org/post", files=form_data)
            # do I need to handle this ? wait for next tick
            logging.error("failed to upload profile: url: %s, code: %s, %s, %s",
                          res.request.url, res.status_code, res.reason, res.text)

        self.reset()
        self._rwlock.runlock()

    def reset(self):
        self._params['from'] = None  # reset the start time
        self._buf.close()
        self._buf = io.StringIO()

import gzip
import re
import io
from dataclasses import dataclass, field
from enum import Enum
from typing import Any, Dict, Tuple, Optional, List

from .profile_pb2 import Profile


class Mode(Enum):
    """profiling mode."""

    CPU = 0
    WALL = 1
    MEMORY = 2
    FULL = 3

    @classmethod
    def from_metadata(cls, mode: str) -> "Mode":
        """Get mode from metadata information."""
        return {
            "cpu": Mode.CPU,
            "wall": Mode.WALL,
            "memory": Mode.MEMORY,
            "full": Mode.FULL,
        }[mode]

class MetricType(Enum):
    """Sample metric type."""

    TIME = 0
    MEMORY = 1

    @classmethod
    def from_mode(cls, mode: str) -> Optional["MetricType"]:
        """Convert metadata mode to metric type."""
        return {
            "cpu": MetricType.TIME,
            "wall": MetricType.TIME,
            "memory": MetricType.MEMORY,
            "full": None,
        }.get(mode)




@dataclass(frozen=True)
class Metric:
    """Austin metrics."""

    type: MetricType
    value: int = 0

    def __add__(self, other: "Metric") -> "Metric":
        """Add metrics together (algebraically)."""
        assert self.type == other.type
        return Metric(
            type=self.type,
            value=self.value + other.value,
        )

    def __sub__(self, other: "Metric") -> "Metric":
        """Subtract metrics (algebraically)."""
        assert self.type == other.type
        return Metric(
            type=self.type,
            value=self.value - other.value,
        )

    def __gt__(self, other: "Metric") -> bool:
        """Strict comparison of metrics."""
        assert self.type == other.type
        return self.value > other.value

    def __ge__(self, other: "Metric") -> bool:
        """Comparison of metrics."""
        assert self.type == other.type
        return self.value >= other.value

    def copy(self) -> "Metric":
        """Make a copy of this object."""
        return dataclasses.replace(self)

    @staticmethod
    def parse(metrics: str, metric_type: Optional[MetricType] = None) -> List["Metric"]:
        """Parse the metrics from a sample.

        Returns a tuple containing the parsed metrics and the head of the
        sample for further processing.
        """
        try:
            ms = [int(_) for _ in metrics.split(",")]
            if len(ms) == 3:
                return [
                    Metric(MetricType.TIME, ms[0] if ms[1] == 0 else 0),
                    Metric(MetricType.TIME, ms[0]),
                    Metric(MetricType.MEMORY, ms[2] if ms[2] >= 0 else 0),
                    Metric(MetricType.MEMORY, -ms[2] if ms[2] < 0 else 0),
                ]
            elif len(ms) != 1:
                raise ValueError()

            assert metric_type is not None

            return [Metric(metric_type, ms[0])]

        except ValueError:
            raise Exception(metrics) from None

    def __str__(self) -> str:
        """Stringify the metric."""
        return str(self.value)


@dataclass(frozen=True)
class Frame:
    """Python frame."""

    function: str
    filename: str
    line: int = 0

    @staticmethod
    def parse(frame: str) -> "Frame":
        """Parse the given string as a frame.

        A string representing a frame has the structure

            ``[frame] := <module>:<function>:<line number>``

        This static method attempts to parse the given string in order to
        identify the parts of the frame and returns an instance of the
        :class:`Frame` dataclass with the corresponding fields filled in.
        """
        if not frame:
            raise Exception(frame)

        try:
            module, function, line = frame.rsplit(":", maxsplit=2)
        except ValueError:
            raise Exception(frame) from None
        return Frame(function, module, int(line))

    def __str__(self) -> str:
        """Stringify the ``Frame`` object."""
        return f"{self.filename}:{self.function}:{self.line}"



@dataclass
class Sample:
    """Austin sample."""

    pid: int
    thread: str
    metric: Metric
    frames: List[Frame] = field(default_factory=list)

    _ALT_FORMAT_RE = re.compile(r";L([0-9]+)")

    @staticmethod
    def is_full(sample: str) -> bool:
        """Determine whether the sample has full metrics."""
        try:
            _, _, metrics = sample.rpartition(" ")
            return len(metrics.split(",")) == 3
        except (ValueError, IndexError):
            return False

    @staticmethod
    def parse(sample: str, metric_type: Optional[MetricType] = None) -> List["Sample"]:
        """Parse the given string as a frame.

        A string representing a sample has the structure

            ``P<pid>;T<tid>[;[frame]]* [metric][,[metric]]*``

        This static method attempts to parse the given string in order to
        identify the parts of the sample and returns an instance of the
        :class:`Sample` dataclass with the corresponding fields filled in.
        """
        if not sample:
            raise Exception(sample)

        if sample[0] != "P":
            raise Exception(f"No process ID in sample '{sample}'")

        head, _, metrics = sample.rpartition(" ")
        process, _, rest = head.partition(";")
        try:
            pid = int(process[1:])
        except ValueError:
            raise Exception(f"Invalid process ID in sample '{sample}'") from None

        if rest[0] != "T":
            raise Exception(f"No thread ID in sample '{sample}'")

        thread, _, frames = rest.partition(";")
        thread = thread[1:]

        if frames:
            if frames.rfind(";L"):
                frames = Sample._ALT_FORMAT_RE.sub(r":\1", frames)

        try:
            ms = Metric.parse(metrics, metric_type)
            return [
                Sample(
                    pid=int(pid),
                    thread=thread,
                    metric=metric,
                    frames=[Frame.parse(frame) for frame in frames.split(";")] if frames else [],
                )
                for metric in ms
            ]
        except Exception as e:
            raise e





@dataclass(frozen=True)
class Location:
    function_id: int
    line_number: int


@dataclass(frozen=True)
class Function:
    name: str
    filename: str


class Producer:

    """Producer generates the profiling file respect to the pprof format
    samples = [...]
    p = Producer()
    p.create_profile_from_samples(samples, mode, rate, 1)
    p.output()
    """

    def __init__(self, mode: str):
        self._string_table: Dict[str, int] = {}
        self._location_map: Dict[Frame, int] = {}
        self._function_map: Dict[Tuple[str, str], int] = {}

        # Create the protobuf Profile message
        self.profile = Profile()
        self.profile.string_table.append("")

        m = Mode.from_metadata(mode)

        if m == Mode.MEMORY:
            self._add_memory_sample_types()

        elif m == Mode.CPU:
            self._add_time_sample_type("cpu")

        elif m == Mode.WALL:
            self._add_time_sample_type("wall")

        elif m == Mode.FULL:
            self._add_time_sample_type("cpu")
            self._add_time_sample_type("wall")
            self._add_memory_sample_types()


    def get_string(self, string: str) -> int:
        """Get the string table index for the given string."""
        try:
            return self._string_table[string]
        except KeyError:
            index = len(self.profile.string_table)
            self._string_table[string] = index
            self.profile.string_table.append(string)
            return index

    def output(self):
        """produce the serialized and gzip-compressed profile proto format."""
        p = self.profile.SerializeToString()
        out = io.BytesIO()
        with gzip.GzipFile(fileobj=out, mode='wb') as f:
            f.write(p)

        return out.getvalue()

    def outputio(self):
        """produce the serialized and gzip-compressed profile proto format."""
        p = self.profile.SerializeToString()
        out = io.BytesIO()
        with gzip.GzipFile(fileobj=out, mode='wb') as f:
            f.write(p)

        out.seek(0)
        return out

    def get_function(self, frame: Frame) -> int:
        """Get the function id from the given Austin frame."""
        key = (frame.function, frame.filename)
        try:
            return self._function_map[key]
        except KeyError:
            function = self.profile.function.add()
            function.id = len(self.profile.function)
            function.name = self.get_string(frame.function)
            function.filename = self.get_string(frame.filename)

            self._function_map[key] = function.id

            return function.id

    def get_location(self, frame: Frame) -> int:
        """Get the location id from the given Austin frame."""
        try:
            return self._location_map[frame]
        except KeyError:
            location = self.profile.location.add()
            location.id = len(self.profile.location)
            line = location.line.add()
            line.function_id = self.get_function(frame)
            line.line = frame.line

            self._location_map[frame] = location.id

            return location.id

    def _add_sample_type(self, type: str, unit: str) -> None:
        _ = self.profile.sample_type.add()
        _.type = self.get_string(type)
        _.unit = self.get_string(unit)

    def _add_time_sample_type(self, type: str) -> None:
        self._add_sample_type(type, "nanoseconds")

    def _add_memory_sample_types(self) -> None:
        self._add_sample_type("allocations", "bytes")
        self._add_sample_type("deallocations", "bytes")

    def add_label_to_sample(self, sample: Sample, key: Any, value: Any) -> None:
        """Add a sample label to the given sample.

        The ``key`` and ``value`` arguments are both converted to strings.
        """
        _ = sample.label.add()
        _.key = self.get_string(str(key))
        _.str = self.get_string(str(value))

    def create_profile_from_samples(self, samples_array, period: int, duration_ns: int):
        """create a profile data structure which follows the proto definition."""
        self.profile.period = period
        self.profile.duration_nanos = duration_ns

        for samples in samples_array:

            pprof_sample = self.profile.sample.add()

            # Add metrics
            for sample in samples:
                pprof_sample.value.append(sample.metric.value)

            # Add locations. Top of the stack first.
            for frame in samples[0].frames[::-1]:
                pprof_sample.location_id.append(self.get_location(frame))

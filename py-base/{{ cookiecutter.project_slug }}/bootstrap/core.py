import abc
import logging
from typing import Any, Dict, List, Optional, Type

logger = logging.getLogger(__name__)


class AlreadyRegistered(Exception):
    pass


_c: Dict = {}


class Container:
    def __init__(self):
        Container.set_context({})

    @staticmethod
    def get_context():
        return _c

    @staticmethod
    def set_context(value: Any):
        global _c
        _c = value

    @classmethod
    def clear(cls):
        cls.set_context({})

    @classmethod
    def _register(cls, id: str, obj: Any):
        mapping = cls.get_context()  # type: Dict[str, Any]

        if id in mapping:
            raise AlreadyRegistered(f"{id} was already registered")

        mapping[id] = obj

    @classmethod
    def _unregister(cls, id: str):
        mapping = cls.get_context()  # type: Dict[str, Any]
        if id in mapping:
            del mapping[id]

    @classmethod
    def register(cls, m: "Manager", obj: Any):
        cls._register(m.id, obj)

    @classmethod
    def unregister(cls, m: "Manager"):
        cls._unregister(m.id)

    @classmethod
    def get(cls, id: str) -> Any:
        mapping = cls.get_context()  # type: Dict[str, Any]
        if id in mapping:
            return mapping[id]

        return None


class Manager(abc.ABC):

    """Manager handles the resource's life cycle and the app flow of process

    Every subclass should be given a name to identify what it handles.
    """

    def __init__(self):
        pass

    @property
    @abc.abstractmethod
    def id(self) -> str:
        """used as the key for the container map"""
        raise NotImplementedError

    @abc.abstractmethod
    def start(self, container: Container):
        raise NotImplementedError

    @abc.abstractmethod
    def close(self, container: Container):
        raise NotImplementedError


class Bootstrap:
    def __init__(self):
        self.successful_managers: Dict[str, Manager] = {}
        self.container = Container()

    def get_manager(self, id: str) -> Optional[Manager]:
        return self.successful_managers.get(id)

    def release(self):
        for _, m in self.successful_managers.items():
            m.close(self.container)

        # all manager are done of resource release so reset the successful_managers
        self.successful_managers = {}

    def boot(self, executor: List[Type[Manager]]):
        """To execute the pre process before starting the application.
        The pre process may be something like db, redis initialization
        or other resources setup.
        """
        for ex in executor:
            try:
                m = ex()
                m.start(self.container)
                self.successful_managers[m.id] = m
                logger.info(f"{m.id} is loaded.")
            except AlreadyRegistered as e:
                logger.warning(e)
            except Exception as e:
                self.release()
                raise e

        logger.info("all settings are loaded.")

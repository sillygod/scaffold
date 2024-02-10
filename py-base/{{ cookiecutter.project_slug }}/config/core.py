from pydantic_settings import BaseSettings

from bootstrap.constants import Identity
from bootstrap.core import Container, Manager


class Settings(BaseSettings):
    log_level: str = "debug"


class ConfigManager(Manager):
    @property
    def id(self) -> str:
        return Identity.CONFIG

    def start(self, container: Container):
        settings = Settings(_env_file='.env')
        container.register(self, settings)

    def close(self, container: Container):
        # no resources need to be released.
        pass

from enum import Enum


class Identity(str, Enum):

    """
    Define the members for managers.
    ex.
    DB = "db"
    """

    CONFIG = "config"

    def __str__(self):
        return self.value

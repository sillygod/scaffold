"""
python shell.py

To enter an interactive shell with loading the resources.
"""

import logging

from bootstrap.core import Bootstrap
from config.core import ConfigManager

available_shells = ["ipython", "python"]


def _ipython():
    from IPython import start_ipython

    bootstrap = Bootstrap()
    bootstrap.boot(
        [
            ConfigManager,
        ]
    )

    logging.getLogger('asyncio.selector_events').propagate = False
    logging.getLogger('parso').propagate = False

    start_ipython(argv=["-i", "shell_config.py"])

    bootstrap.release()


def _python():
    # borrow the following code from the django source code
    import code

    imported_objects = {}

    try:
        import readline
    except ImportError:
        pass
    else:
        import rlcompleter

        readline.set_completer(rlcompleter.Completer(imported_objects).complete)
        readline_doc = getattr(readline, '__doc__', '')
        if readline_doc is not None and 'libedit' in readline_doc:
            readline.parse_and_bind("bind ^I rl_complete")
        else:
            readline.parse_and_bind("tab:complete")

    code.interact(local=imported_objects)


shell_map = {
    'ipython': _ipython,
    'python': _python,
}

if __name__ == "__main__":
    for shell in available_shells:
        try:
            shell_map[shell]()
            break
        except ImportError:
            pass
        except Exception as e:
            logging.error(e)
            raise e

    logging.info("exit.")

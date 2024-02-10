FROM python:3.11.6-slim-bullseye
ADD  . /app
RUN apt update && apt-get install -y procps
RUN pip install poetry
ENV POETRY_VIRTUALENVS_IN_PROJECT=1
RUN cd /app && poetry install

ENV PYTHONUNBUFFERED 1
ENV PYTHONOPTIMIZE 1
ENV PORT 8000
WORKDIR /app
ENV PATH "/app/.venv/bin:$PATH"
CMD ["main.py"]

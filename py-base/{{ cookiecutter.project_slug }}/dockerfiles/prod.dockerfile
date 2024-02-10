FROM python:3.11.6-slim-bullseye as builder
ADD  . /app
RUN apt update && apt-get install -y procps
RUN pip install poetry
ENV POETRY_VIRTUALENVS_IN_PROJECT=1
RUN cd /app && poetry install --without dev --no-root


FROM python:3.11.6-slim-bullseye
COPY --from=builder /app/.venv /app/.venv
COPY --from=builder /bin /bin
COPY --from=builder /lib /lib
COPY . /app

ENV PYTHONUNBUFFERED 1
ENV PYTHONOPTIMIZE 1
ENV PORT 8000

WORKDIR /app

ENV PATH "/app/.venv/bin:$PATH"
CMD ["main.py"]

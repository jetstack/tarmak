# Copyright Jetstack Ltd. See LICENSE for details.
FROM jfloff/alpine-python:3.4-slim

WORKDIR /site

RUN apk update && apk add enchant make

COPY requirements.txt .
RUN pip install -r requirements.txt
RUN mkdir -p venv/bin && ln -s $(which python) venv/bin/python && touch .venv

ENTRYPOINT ["make"]
CMD ["spelling", "linkcheck", "html"]

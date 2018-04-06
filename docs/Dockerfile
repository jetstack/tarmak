# Copyright Jetstack Ltd. See LICENSE for details.
FROM python:3

WORKDIR /site

RUN apt-get update && apt-get install -y python-enchant wbritish

COPY requirements.txt .
RUN pip install -r requirements.txt
RUN mkdir -p venv/bin && ln -s $(which python) venv/bin/python && touch .venv

ENTRYPOINT ["make"]
CMD ["spelling", "linkcheck", "html"]

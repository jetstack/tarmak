# Copyright Jetstack Ltd. See LICENSE for details.
FROM python:3.6-stretch

WORKDIR /site

# ensure python and dependencies are installed
RUN apt-get update && apt-get install -y python-enchant wbritish

# install sphinx
COPY requirements.txt .
ADD generated /site/generated
RUN pip install -r requirements.txt
RUN mkdir -p venv/bin && ln -s $(which python) venv/bin/python && touch .venv

ENTRYPOINT ["make"]
CMD ["spelling", "linkcheck", "html"]

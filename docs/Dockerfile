FROM jfloff/alpine-python:3.4-slim

WORKDIR /site

RUN apk update && apk add enchant make

RUN pip install --upgrade pip virtualenv
RUN virtualenv venv
RUN source venv/bin/activate

COPY requirements.txt .
RUN pip install -r requirements.txt

ENTRYPOINT ["make"]
CMD ["html"]

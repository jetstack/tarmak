FROM node:alpine

WORKDIR /docs
COPY brodocs /docs/.
RUN rm -rf /docs/documents/*
COPY manifest.json /docs/.
COPY includes /docs/documents/

ENTRYPOINT ["node"]
CMD ["brodoc.js", "/docs/manifest.json", "/docs/includes", "/docs/output"]

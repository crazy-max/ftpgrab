# syntax=docker/dockerfile:1

ARG MKDOCS_VERSION="8.3.9"

FROM squidfunk/mkdocs-material:${MKDOCS_VERSION} AS base
RUN apk add --no-cache git git-fast-import openssl \
  && apk add --no-cache --virtual .build gcc musl-dev \
  && pip install --no-cache-dir \
    'lunr==0.7.0' \
    'markdown-include==0.8.1' \
    'mkdocs-awesome-pages-plugin==2.9.3' \
    'mkdocs-exclude==1.0.2' \
    'mkdocs-git-revision-date-localized-plugin==1.3.0' \
    'mkdocs-macros-plugin==1.3.7' \
  && apk del .build gcc musl-dev \
  && rm -rf /tmp/*

FROM base AS generate
RUN --mount=type=bind,target=. \
  mkdocs build --strict --site-dir /out

FROM scratch AS release
COPY --from=generate /out /

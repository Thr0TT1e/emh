FROM registry.altlinux.org/p11/postgresql:latest

USER root

ENV PGDATA=/var/lib/pgsql/data

RUN mkdir -p /docker-entrypoint-initdb.d

COPY ./docker-entrypoint-initdb.d /docker-entrypoint-initdb.d

RUN apt-get update && apt-get install -y gosu

USER postgres

COPY ./docker-entrypoint.sh /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]

EXPOSE 5432

CMD ["postgres"]

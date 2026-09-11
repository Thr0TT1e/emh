#!/usr/bin/env bash
set -e

file_env() {
	local var="$1"
	local fileVar="${var}_FILE"
	local def="${2:-}"
	if [ "${!var:-}" ] && [ "${!fileVar:-}" ]; then
		echo >&2 "error: both $var and $fileVar are set (but are exclusive)"
		exit 1
	fi
	local val="$def"
	if [ "${!var:-}" ]; then
		val="${!var}"
	elif [ "${!fileVar:-}" ]; then
		val="$(< "${!fileVar}")"
	fi
	export "$var"="$val"
	unset "$fileVar"
}

if [ "${1:0:1}" = '-' ]; then
	set -- postgres "$@"
fi

export PGDATA="/var/lib/pgsql/data"

echo "разрешить запуск контейнера с помощью --user"

if [ "$1" = 'postgres' ] && [ "$(id -u)" = '0' ]; then
	mkdir -p "$PGDATA"
	chown -R postgres "$PGDATA"
	chmod 700 "$PGDATA"

	mkdir -p /var/run/postgresql
	chown -R postgres /var/run/postgresql
	chmod 775 /var/run/postgresql

	echo "Создайте каталог журнала транзакций перед запуском initdb (см. ниже), чтобы каталог принадлежал нужному пользователю"
	if [ "$POSTGRES_INITDB_WALDIR" ]; then
		mkdir -p "$POSTGRES_INITDB_WALDIR"
		chown -R postgres "$POSTGRES_INITDB_WALDIR"
		chmod 700 "$POSTGRES_INITDB_WALDIR"
	fi

	exec gosu postgres "$BASH_SOURCE" "$@"
fi

if [ "$1" = 'postgres' ]; then
	mkdir -p "$PGDATA"
	chown -R "$(id -u)" "$PGDATA" 2>/dev/null || :
	chmod 700 "$PGDATA" 2>/dev/null || :

	echo "обратите особое внимание на PG_VERSION, как это ожидается в каталоге базы данных"
	if [ ! -s "$PGDATA/PG_VERSION" ]; then
		file_env 'POSTGRES_INITDB_ARGS'
		if [ "$POSTGRES_INITDB_WALDIR" ]; then
			export POSTGRES_INITDB_ARGS="$POSTGRES_INITDB_ARGS --waldir $POSTGRES_INITDB_WALDIR"
		fi
		eval "initdb --username=postgres $POSTGRES_INITDB_ARGS --encoding=UTF8 --locale=ru_RU.UTF-8 --lc-collate=ru_RU.UTF-8 --lc-ctype=ru_RU.UTF-8 --lc-messages=C"

		echo "сначала проверьте пароль, чтобы мы могли вывести предупреждение до того, как postgres все испортит"

		file_env 'DB_NAME'
		file_env 'DB_USER'
		file_env 'DB_PASSWORD'

		echo "DB_NAME: $DB_NAME DB_PASSWORD: $DB_PASSWORD DB_USER: $DB_USER"

		if [ "$DB_PASSWORD" ]; then
			pass="PASSWORD '$DB_PASSWORD'"
			authMethod=md5
		else
			echo "Параметр - запрещает вводные символы табуляции, но не пробелы. :)"
			cat >&2 <<-'EOWARN'
				****************************************************
				WARNING: No password has been set for the database.
				         This will allow anyone with access to the
				         Postgres port to access your database. In
				         Docker's default configuration, this is
				         effectively any other container on the same
				         system.

				         Use "-e DB_PASSWORD=password" to set
				         it in "docker run".
				****************************************************
			EOWARN

			pass=
			authMethod=trust
		fi

		{
			echo
			echo "host all all all $authMethod"
		} >> "$PGDATA/pg_hba.conf"


		sed -i '/listen_addresses = \x27localhost\x27/c\listen_addresses = \x27*\x27' $PGDATA/postgresql.conf

		# internal start of server in order to allow set-up using psql-client
		# does not listen on external TCP/IP and waits until start finishes
		PGUSER="${PGUSER:-postgres}" \
		pg_ctl -D "$PGDATA" \
		    -o "-c listen_addresses='*'" \
			-w start

		file_env 'DB_USER' 'postgres'
		file_env 'DB_NAME' "$DB_USER"

		psql=( psql -v ON_ERROR_STOP=1 )

		if [ "$DB_NAME" != 'postgres' ]; then
         		echo "CREATE DATABASE $DB_NAME"
			"${psql[@]}" --username postgres <<-EOSQL
                                CREATE DATABASE "$DB_NAME" ENCODING 'UTF8' LC_COLLATE 'ru_RU.UTF-8'  LC_CTYPE 'ru_RU.UTF-8' TEMPLATE template0;
			EOSQL
			echo
		fi

		if [ "$DB_USER" = 'postgres' ]; then
			op='ALTER'
		else
			op='CREATE'
		fi
		"${psql[@]}" --username postgres <<-EOSQL
			$op USER "$DB_USER" WITH SUPERUSER $pass ;
		EOSQL
		echo

		psql+=( --username "$DB_USER" --dbname "$DB_NAME" )

		echo

                echo "Проверка наличия скрипта make_deploy.sh"

                if [ -x /docker-entrypoint-initdb.d/make_deploy.sh ]; then
                    echo "Бегущий make_deploy.sh..."
                    /docker-entrypoint-initdb.d/make_deploy.sh || {
                        echo "make_deploy.sh ошибка выполнения!" >&2
                        exit 1
                    }
                else
                    echo "make_deploy.sh файл не найден или не исполняется, пропускается."
                fi

                for f in /docker-entrypoint-initdb.d/*; do
                    case "$f" in
                        *.sh)
                            [ "$f" = "/docker-entrypoint-initdb.d/make_deploy.sh" ] && continue
                            echo "$0: running $f"; . "$f" ;;
                        *.sql)
                            echo "$0: running $f"; "${psql[@]}" -f "$f"; echo ;;
                        *.sql.gz)
                            echo "$0: running $f"; gunzip -c "$f" | "${psql[@]}"; echo ;;
                        *)
                            echo "$0: ignoring $f" ;;
                    esac
                    echo
                done

		PGUSER="${PGUSER:-postgres}" \
		pg_ctl -D "$PGDATA" -m fast -w stop

		echo
		echo 'Процесс инициализации PostgreSQL завершен; готов к запуску.'
		echo
	fi
fi

exec "$@"

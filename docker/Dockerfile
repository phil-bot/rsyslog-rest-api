FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=Europe/Berlin

# Install rsyslog, MariaDB, and basic utilities
RUN apt-get update && apt-get install -y \
    rsyslog \
    rsyslog-mysql \
    mariadb-server \
    mariadb-client \
    curl \
    && rm -rf /var/lib/apt/lists/*

# MariaDB setup
RUN mkdir -p /var/run/mysqld /var/lib/mysql && \
    chown -R mysql:mysql /var/run/mysqld /var/lib/mysql && \
    mysql_install_db --user=mysql --datadir=/var/lib/mysql

# Create directories
RUN mkdir -p /opt/rsyslog-rest-api

# Copy entrypoint
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 8000
WORKDIR /root

ENTRYPOINT ["/entrypoint.sh"]
CMD ["tail", "-f", "/dev/null"]

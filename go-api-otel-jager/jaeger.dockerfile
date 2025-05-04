FROM jaegertracing/all-in-one:latest
#Workaround to create folders we must use root
USER root
RUN mkdir -p /var/badger/{key,data} && \
    # chown -R nobody:nobody /var/badger && \
    chmod -R 777 /var/badger
# USER nobody  # Switch back to the non-root user
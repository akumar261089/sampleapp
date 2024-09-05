# Use scratch image
FROM scratch

# Add application binary to the container
COPY webserver /webserver
COPY templates /templates

COPY styles.css /styles.css

# Run the application
ENTRYPOINT ["/webserver"]

# Use a CentOS-based Java image as the base
FROM fabric8/java-centos-openjdk11-jre

# Add the Filebeat tarball
ADD https://artifacts.elastic.co/downloads/beats/filebeat/filebeat-8.9.1-linux-x86_64.tar.gz /tmp/filebeat.tar.gz

# Switch to root user to perform installation
USER root

# Extract the Filebeat tarball to the temporary directory and remove it afterward
RUN tar -zxvf /tmp/filebeat.tar.gz -C /tmp/ && rm -rf /tmp/filebeat.tar.gz
RUN mv /tmp/filebeat-* /usr/share/filebeat

# Copy the Filebeat configuration and set permissions
COPY filebeat.yml /usr/share/filebeat/filebeat.yml
RUN chmod 600 /usr/share/filebeat/filebeat.yml

# Set the working directory
WORKDIR /usr/share/filebeat


# Define the entrypoint for running Filebeat
CMD ["/usr/share/filebeat/filebeat", "-e", "-c", "/usr/share/filebeat/filebeat.yml"]

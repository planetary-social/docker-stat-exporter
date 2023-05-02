# Docker Stat Exporter

This is a small, basic metrics exporter for prometheus that is set to answer the question "Are my docker containers writing logs"?

It uses the official docker client for go to poll your docker daemon, get the logs for the last 12 hours, and count the # of entries per container.

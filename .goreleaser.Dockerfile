FROM gcr.io/distroless/static-debian12:nonroot
COPY wolproxy /wolproxy
ENTRYPOINT ["/wolproxy"]
CMD ["run"]
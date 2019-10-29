FROM gcr.io/gcpug-container/appengine-go:1.11
COPY ./land /land
ENTRYPOINT ["/land"]
FROM alpine:3.6 as alpine
RUN apk add -U --no-cache ca-certificates && apk add -U --no-cache bash
ENV PATH="/:${PATH}"
COPY ccloud ccloud-* /
ENTRYPOINT ["/bin/bash"]

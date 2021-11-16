FROM golang:1.13-alpine AS build
RUN apk add --no-cache gcc libc-dev && mkdir /pingu
WORKDIR /pingu
COPY go.mod go.sum ./
RUN go mod download
COPY pingu/ pingu/
COPY plugins/ plugins/
ARG SOURCE_COMMIT
RUN for d in plugins/*/ ; do \
        echo "${d}" \
        go build -v -trimpath -ldflags "-X main.version=${SOURCE_COMMIT}" -buildmode=plugin -o plugins/$(basename $d).so $d*; \
    done
COPY pingu.go .
RUN BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` && go build -trimpath -v -ldflags "-X github.com/jyggen/pingu/pingu.builtAt=${BUILD_DATE} -X github.com/jyggen/pingu/pingu.version=${SOURCE_COMMIT}" -o bin/pingu pingu.go && chmod +x bin/pingu

FROM alpine
RUN mkdir /pingu /pingu/plugins
WORKDIR /pingu
COPY --from=build /pingu/bin/pingu pingu
COPY --from=build /pingu/plugins/*.so ./plugins/
ENV AOC_TIMEOUT=5 JIRA_TIMEOUT=5 PINGU_PLUGIN_PATH=/pingu/plugins CRON_TZ=UTC
CMD ["/pingu/pingu"]

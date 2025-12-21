# Собираем в гошке
FROM golang:1.22 as build

ARG BIN_FILE
ENV BIN_FILE_ENV /opt/calendar/${BIN_FILE}
ENV CODE_DIR /go/src/

WORKDIR ${CODE_DIR}

# Кэшируем слои с модулями
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . ${CODE_DIR}

ARG LDFLAGS
ARG CMD_PATH
RUN CGO_ENABLED=0 go build \
        -ldflags "$LDFLAGS" \
        -o ${BIN_FILE_ENV} ${CMD_PATH}

FROM alpine:3.18

LABEL ORGANIZATION="OTUS Online Education"
LABEL SERVICE="calendar"
LABEL MAINTAINERS="student@otus.ru"

ARG BIN_FILE
ENV BIN_FILE_ENV "/opt/calendar/${BIN_FILE}"
COPY --from=build ${BIN_FILE_ENV} ${BIN_FILE_ENV}

ENV CONFIG_FILE /etc/calendar/config.yaml
COPY ./configs/config.yaml ${CONFIG_FILE}

ENTRYPOINT ["sh", "-c", "${BIN_FILE_ENV} -config ${CONFIG_FILE}"]

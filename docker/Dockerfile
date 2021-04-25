FROM docker:20-git
LABEL maintainer="simone.bembi@gmail.com"

COPY ./release/polo-linux-amd64 /app/polo
WORKDIR /app
RUN chmod +x polo

ENTRYPOINT ./polo
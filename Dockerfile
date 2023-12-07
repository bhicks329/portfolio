FROM golang
# create a working directory
WORKDIR /app
# add source code
ADD src .
# run main.go
CMD ["go", "run", "main.go"]
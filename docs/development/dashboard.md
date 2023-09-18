# Dashboard
The Dashboard service provides a web interface for users

## Configuration

| Variable                    | Description                                                                                           | Required | Default                   |
| --------------------------- | ----------------------------------------------------------------------------------------------------- | -------- | ------------------------- |
| HTTP_ADDR                   | HTTP Address on which to bind the devices, measurements and pipeline APIs                             | no       | :3000                     |
| GO_ENV                   | Whether running in production or development environment. This influences if live-reload and static files are served | no       | Development |
| STATIC_PATH                   | The relative path where the static files are served from | yes if GO_ENV is development       |                     |

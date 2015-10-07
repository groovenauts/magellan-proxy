## 0.1.0

- Add support for MQTT Publish message for worker.
  `PUBLISH` message was translated into HTTP POST request to web server worker.
  Topic is passed via query parameter `topic`, and payload become request body.
  You can specify path to POST request with `--publish` command line option.

## 0.0.2

- exit magellan-proxy on connection loss with TRMQ.
  MAGELLAN will automatically reboot workers on such cases.

## 0.0.1

- Initial Release

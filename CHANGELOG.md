## 0.1.3

- Wait until backend application server start to listen socket before
  dispatching requests.

## 0.1.2

- Just rebuild with Go 1.5.3 for fix potential vulnerability.
  see https://groups.google.com/forum/#!topic/golang-nuts/MEATuOi_ei4

## 0.1.1

- Shrink message channel capacity.
  magellan-proxy subscribe AMQP and messages are dispatched by MQ in front of magellan-proxy.
  magellan-proxy should not prefetch and queueing messages for http worker.

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

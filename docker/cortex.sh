#!/bin/sh
# cortexboard `start` solo escucha en 127.0.0.1. Dentro del contenedor se publica
# por socat en 0.0.0.0:4749 para que el host y los demás servicios lleguen.
cortexboard start --port 4748 &
socat TCP-LISTEN:4749,fork,reuseaddr TCP:127.0.0.1:4748

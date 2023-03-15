package xyz.block.ftl.control

import java.net.InetSocketAddress
import java.net.SocketAddress
import java.net.UnixDomainSocketAddress

/**
 * Parses a socket URI in the form unix://PATH or tcp://HOST:PORT
 */
fun parseSocket(socket: String): SocketAddress? {
  if (socket == "") return null
  val schema = socket.split("://", limit = 2)
  return when (schema[0]) {
    //  TODO(aat) Remove this.
    //   Unix sockets are effectively unusable in gRPC, unfortunately, as
    //   they require either kqueue or epoll native implementations
    "unix" -> UnixDomainSocketAddress.of(schema[1])
    "tcp" -> {
      val hostPort = schema[1].split(":", limit = 2)
      return InetSocketAddress(hostPort[0], hostPort[1].toInt())
    }

    else -> throw RuntimeException("unsupported socket type ${schema[0]}")
  }
}
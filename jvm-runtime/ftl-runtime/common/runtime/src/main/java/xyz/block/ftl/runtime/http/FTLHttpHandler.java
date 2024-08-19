package xyz.block.ftl.runtime.http;

import java.io.ByteArrayOutputStream;
import java.net.InetSocketAddress;
import java.nio.channels.Channels;
import java.nio.channels.WritableByteChannel;
import java.nio.charset.StandardCharsets;
import java.util.Base64;
import java.util.HashMap;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.Set;
import java.util.concurrent.CompletableFuture;

import jakarta.inject.Singleton;
import jakarta.ws.rs.core.MediaType;

import org.jboss.logging.Logger;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.protobuf.ByteString;

import io.netty.buffer.ByteBuf;
import io.netty.buffer.Unpooled;
import io.netty.channel.FileRegion;
import io.netty.handler.codec.http.DefaultHttpRequest;
import io.netty.handler.codec.http.DefaultLastHttpContent;
import io.netty.handler.codec.http.HttpContent;
import io.netty.handler.codec.http.HttpHeaderNames;
import io.netty.handler.codec.http.HttpMethod;
import io.netty.handler.codec.http.HttpResponse;
import io.netty.handler.codec.http.HttpVersion;
import io.netty.handler.codec.http.LastHttpContent;
import io.netty.util.ReferenceCountUtil;
import io.quarkus.netty.runtime.virtual.VirtualClientConnection;
import io.quarkus.netty.runtime.virtual.VirtualResponseHandler;
import io.quarkus.vertx.http.runtime.QuarkusHttpHeaders;
import io.quarkus.vertx.http.runtime.VertxHttpRecorder;
import xyz.block.ftl.runtime.FTLRecorder;
import xyz.block.ftl.runtime.builtin.HttpRequest;
import xyz.block.ftl.v1.CallRequest;
import xyz.block.ftl.v1.CallResponse;

@SuppressWarnings("unused")
@Singleton
public class FTLHttpHandler {

    public static final String CONTENT_TYPE = "Content-Type";
    final ObjectMapper mapper;
    private static final Logger log = Logger.getLogger("quarkus.amazon.lambda.http");

    private static final int BUFFER_SIZE = 8096;

    private static final Map<String, List<String>> ERROR_HEADERS = Map.of();

    private static final String COOKIE_HEADER = "Cookie";

    // comma headers for headers that have comma in value and we don't want to split it up into
    // multiple headers
    private static final Set<String> COMMA_HEADERS = Set.of("access-control-request-headers");

    public FTLHttpHandler(ObjectMapper mapper) {
        this.mapper = mapper;
    }

    public CallResponse handle(CallRequest in, boolean base64Encoded) {
        try {
            var body = mapper.createParser(in.getBody().newInput())
                    .readValueAs(xyz.block.ftl.runtime.builtin.HttpRequest.class);
            body.getHeaders().put(FTLRecorder.X_FTL_VERB, List.of(in.getVerb().getName()));
            var ret = handleRequest(body, base64Encoded);
            if (ret.getBody() == null) {
                ret.setBody("{}");
            }
            ret.getHeaders().remove("content-length");
            var mappedResponse = mapper.writer().writeValueAsBytes(ret);
            return CallResponse.newBuilder().setBody(ByteString.copyFrom(mappedResponse)).build();
        } catch (Exception e) {
            return CallResponse.newBuilder().setError(CallResponse.Error.newBuilder().setMessage(e.getMessage()).build())
                    .build();
        }

    }

    public xyz.block.ftl.runtime.builtin.HttpResponse handleRequest(HttpRequest request, boolean base64Encoded) {
        InetSocketAddress clientAddress = null;
        try {
            return nettyDispatch(clientAddress, request, base64Encoded);
        } catch (Exception e) {
            log.error("Request Failure", e);
            xyz.block.ftl.runtime.builtin.HttpResponse res = new xyz.block.ftl.runtime.builtin.HttpResponse();
            res.setStatus(500);
            res.setError(e);
            res.setHeaders(ERROR_HEADERS);
            return res;
        }

    }

    private class NettyResponseHandler implements VirtualResponseHandler {
        xyz.block.ftl.runtime.builtin.HttpResponse responseBuilder = new xyz.block.ftl.runtime.builtin.HttpResponse();
        final boolean base64Encoded;
        ByteArrayOutputStream baos;
        WritableByteChannel byteChannel;
        final xyz.block.ftl.runtime.builtin.HttpRequest request;
        CompletableFuture<xyz.block.ftl.runtime.builtin.HttpResponse> future = new CompletableFuture<>();
        boolean json = false;

        public NettyResponseHandler(boolean base64Encoded, xyz.block.ftl.runtime.builtin.HttpRequest request) {
            this.base64Encoded = base64Encoded;
            this.request = request;
        }

        public CompletableFuture<xyz.block.ftl.runtime.builtin.HttpResponse> getFuture() {
            return future;
        }

        @Override
        public void handleMessage(Object msg) {
            try {
                //log.info("Got message: " + msg.getClass().getName());

                if (msg instanceof HttpResponse) {
                    HttpResponse res = (HttpResponse) msg;
                    responseBuilder.setStatus(res.status().code());

                    final Map<String, List<String>> headers = new HashMap<>();
                    responseBuilder.setHeaders(headers);
                    for (String name : res.headers().names()) {
                        final List<String> allForName = res.headers().getAll(name);
                        if (allForName == null || allForName.isEmpty()) {
                            continue;
                        }
                        headers.put(name, allForName);
                    }
                    if (res.headers().contains(CONTENT_TYPE)) {
                        String contentType = res.headers().get(CONTENT_TYPE);
                        if (contentType != null && !contentType.isEmpty()) {
                            json = contentType.toLowerCase(Locale.ROOT).contains("application/json");
                        }
                    }
                }
                if (msg instanceof HttpContent) {
                    HttpContent content = (HttpContent) msg;
                    int readable = content.content().readableBytes();
                    if (baos == null && readable > 0) {
                        baos = createByteStream();
                    }
                    for (int i = 0; i < readable; i++) {
                        baos.write(content.content().readByte());
                    }
                }
                if (msg instanceof FileRegion) {
                    FileRegion file = (FileRegion) msg;
                    if (file.count() > 0 && file.transferred() < file.count()) {
                        if (baos == null)
                            baos = createByteStream();
                        if (byteChannel == null)
                            byteChannel = Channels.newChannel(baos);
                        file.transferTo(byteChannel, file.transferred());
                    }
                }
                if (msg instanceof LastHttpContent) {
                    if (baos != null) {
                        if (json) {
                            responseBuilder.setBody(baos.toString(StandardCharsets.UTF_8));
                        } else if (base64Encoded) {
                            responseBuilder.setBody(
                                    mapper.writer().writeValueAsString(Base64.getEncoder().encodeToString(baos.toByteArray())));
                        } else {
                            responseBuilder.setBody(mapper.writer().writeValueAsString(baos.toString(StandardCharsets.UTF_8)));
                        }
                        List<String> ct = responseBuilder.getHeaders().get(CONTENT_TYPE);
                        if (ct == null || ct.isEmpty()) {
                            //TODO: how to handle this
                            responseBuilder.setBody(baos.toString(StandardCharsets.UTF_8));
                        } else if (ct.get(0).contains(MediaType.TEXT_PLAIN)) {
                            // need to encode as JSON string
                        } else {
                        }
                    }
                    future.complete(responseBuilder);
                }
            } catch (Throwable ex) {
                future.completeExceptionally(ex);
            } finally {
                if (msg != null) {
                    ReferenceCountUtil.release(msg);
                }
            }
        }

        @Override
        public void close() {
            if (!future.isDone())
                future.completeExceptionally(new RuntimeException("Connection closed"));
        }
    }

    private xyz.block.ftl.runtime.builtin.HttpResponse nettyDispatch(InetSocketAddress clientAddress,
            HttpRequest request, boolean base64Encoded)
            throws Exception {
        QuarkusHttpHeaders quarkusHeaders = new QuarkusHttpHeaders();
        quarkusHeaders.setContextObject(xyz.block.ftl.runtime.builtin.HttpRequest.class, request);
        HttpMethod httpMethod = HttpMethod.valueOf(request.getMethod());
        if (httpMethod == null) {
            throw new IllegalStateException("Missing HTTP method in request event");
        }
        //TODO: encoding schenanigans
        StringBuilder path = new StringBuilder(request.getPath());
        if (request.getQuery() != null && !request.getQuery().isEmpty()) {
            path.append("?");
            var first = true;
            for (var entry : request.getQuery().entrySet()) {
                for (var val : entry.getValue()) {
                    if (first) {
                        first = false;
                    } else {
                        path.append("&");
                    }
                    path.append(entry.getKey()).append("=").append(val);
                }
            }
        }
        DefaultHttpRequest nettyRequest = new DefaultHttpRequest(HttpVersion.HTTP_1_1,
                httpMethod, path.toString(), quarkusHeaders);
        if (request.getHeaders() != null) {
            for (Map.Entry<String, List<String>> header : request.getHeaders().entrySet()) {
                if (header.getValue() != null) {
                    for (String val : header.getValue()) {
                        nettyRequest.headers().add(header.getKey(), val);
                    }
                }
            }
        }
        nettyRequest.headers().add(CONTENT_TYPE, MediaType.APPLICATION_JSON);

        if (!nettyRequest.headers().contains(HttpHeaderNames.HOST)) {
            nettyRequest.headers().add(HttpHeaderNames.HOST, "localhost");
        }

        HttpContent requestContent = LastHttpContent.EMPTY_LAST_CONTENT;
        if (request.getBody() != null) {
            // See https://www.w3.org/Protocols/rfc2616/rfc2616-sec4.html#sec4.3
            nettyRequest.headers().add(HttpHeaderNames.TRANSFER_ENCODING, "chunked");
            if (base64Encoded) {
                requestContent = new DefaultLastHttpContent(
                        Unpooled.copiedBuffer(Base64.getDecoder().decode(request.getBody().asText())));
            } else if (request.getBody().isTextual()) {
                requestContent = new DefaultLastHttpContent(
                        Unpooled.copiedBuffer(request.getBody().asText(), StandardCharsets.UTF_8));
            } else if (request.getBody().isBigDecimal() || request.getBody().isDouble() || request.getBody().isFloat()
                    || request.getBody().isInt() || request.getBody().isBigInteger()) {
                requestContent = new DefaultLastHttpContent(
                        Unpooled.copiedBuffer(request.getBody().toString(), StandardCharsets.UTF_8));
            } else {
                ByteBuf body = Unpooled.copiedBuffer(request.getBody().toString(), StandardCharsets.UTF_8); //TODO: do we need to look at the request encoding?
                requestContent = new DefaultLastHttpContent(body);
            }
        }
        NettyResponseHandler handler = new NettyResponseHandler(base64Encoded, request);
        VirtualClientConnection connection = VirtualClientConnection.connect(handler, VertxHttpRecorder.VIRTUAL_HTTP,
                clientAddress);

        connection.sendMessage(nettyRequest);
        connection.sendMessage(requestContent);
        try {
            return handler.getFuture().get();
        } finally {
            connection.close();
        }
    }

    private ByteArrayOutputStream createByteStream() {
        ByteArrayOutputStream baos;
        baos = new ByteArrayOutputStream(BUFFER_SIZE);
        return baos;
    }

    private boolean isBinary(String contentType) {
        if (contentType != null) {
            String ct = contentType.toLowerCase(Locale.ROOT);
            return !(ct.startsWith("text") || ct.contains("json") || ct.contains("xml") || ct.contains("yaml"));
        }
        return false;
    }

}

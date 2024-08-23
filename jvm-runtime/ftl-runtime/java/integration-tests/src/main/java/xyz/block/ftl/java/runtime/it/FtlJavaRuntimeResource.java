/*
* Licensed to the Apache Software Foundation (ASF) under one or more
* contributor license agreements.  See the NOTICE file distributed with
* this work for additional information regarding copyright ownership.
* The ASF licenses this file to You under the Apache License, Version 2.0
* (the "License"); you may not use this file except in compliance with
* the License.  You may obtain a copy of the License at
*
*      http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*/
package xyz.block.ftl.java.runtime.it;

import java.nio.charset.StandardCharsets;

import jakarta.enterprise.context.ApplicationScoped;
import jakarta.ws.rs.Consumes;
import jakarta.ws.rs.POST;
import jakarta.ws.rs.Path;
import jakarta.ws.rs.core.MediaType;

import ftl.echo.EchoClient;
import ftl.echo.EchoRequest;
import xyz.block.ftl.Verb;

@Path("/test")
@ApplicationScoped
public class FtlJavaRuntimeResource {

    @POST
    @Consumes(MediaType.APPLICATION_JSON)
    @Path("/post")
    public String post(Person person) {
        return "Hello " + person.first() + " " + person.last();
    }

    @POST
    @Consumes(MediaType.APPLICATION_JSON)
    @Path("/bytes")
    public String bytesHttp(byte[] data) {
        return "Hello " + new String(data, StandardCharsets.UTF_8);
    }

    /**
     * This is a verb FOOO
     *
     * @param name
     * @param echoClient
     * @return
     */
    @Verb
    public String hello(String name, EchoClient echoClient) {
        return "Hello " + echoClient.call(new EchoRequest().setName(name)).getMessage();
    }

    @Verb
    public void publish(Person person, MyTopic topic) {
        topic.publish(person);
    }

    @Verb
    public byte[] bytes(byte[] bytes) {
        return bytes;
    }
}

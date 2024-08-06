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

import jakarta.enterprise.context.ApplicationScoped;
import jakarta.ws.rs.Consumes;
import jakarta.ws.rs.POST;
import jakarta.ws.rs.core.MediaType;
import xyz.block.ftl.Verb;

@ApplicationScoped
public class FtlJavaRuntimeResource {
    // add some rest methods here

    @POST
    @Consumes(MediaType.APPLICATION_JSON)
    public String post(Person person) {
        return "Hello " + person.getFirst() + " " + person.getLast();
    }

    @Verb
    public String hello() {
        return "Hello ftl-java-runtime";
    }
}

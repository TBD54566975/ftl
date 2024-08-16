package xyz.block.ftl.java.test.database;

import jakarta.transaction.Transactional;

import xyz.block.ftl.Verb;

public class Database {

    @Verb
    @Transactional
    public InsertResponse insert(InsertRequest insertRequest) {
        Request request = new Request();
        request.data = insertRequest.getData();
        request.persist();
        return new InsertResponse();
    }
}

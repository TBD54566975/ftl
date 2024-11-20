package xyz.block.ftl.java.test.database;

import jakarta.transaction.Transactional;

import xyz.block.ftl.Verb;
import java.util.Map;
import java.util.List;

public class Database {

    @Verb
    @Transactional
    public InsertResponse insert(InsertRequest insertRequest) {
        Request request = new Request();
        request.data = insertRequest.getData();
        request.persist();
        return new InsertResponse();
    }

    @Verb
    @Transactional
    public Map<String, String> query() {
        List<Request> requests = Request.listAll();
        return Map.of("data", requests.get(0).data);
    }
}

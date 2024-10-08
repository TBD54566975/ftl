# FTL pubsub example

Run using:
```sh
ftl call pubsub.orderPizza '{"type":"veggie","customer":"bob"}'
```

Expect:
```
info:pubsub: Cooking pizza: {99 veggie bob}
info:pubsub: Delivering pizza: {99 veggie bob}
```
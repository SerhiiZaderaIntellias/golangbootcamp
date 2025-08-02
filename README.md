# golangbootcamp
golang bootcamp


### http://localhost:8080

### RUN

> go run ./cmd/rssreader

### Makefile
> make db-up

> make migrate-up

> make run



### REST API
```
POST /feed
Content-Type: application/json

{
  "url": "https://dou.ua/feed/"
}
```

```
GET /feed?title=chatgpt&description=новий&limit=5&offset=0
```

```
GET http://localhost:8080/feed/3
```

```
DELETE http://localhost:8080/feed/3
```


### FLOW
```
HTTP POST /feed
    ↓
handler.CreateFeed()
    ↓
h.pool.Submit(req.URL)
    ↓
p.jobs <- url (channel)
    ↓
[QUEUE: url1, url2, url3...]
    ↓
case url := <-p.jobs (worker receives)
    ↓
rss.FetchAndParse(url)
    ↓
rss.StoreItems(db, items)
    ↓
✅ DONE!
```
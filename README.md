## urlshortener
### Сервис для сокращения ссылок

API:

* POST /shorten {"url": "\<URL\>"} -> {"key": "\<KEY\>"}
* GET /go/\<KEY\> -> 302

В тело `/shorten` запроса будет передаваться json вида
```
{"url":"https://github.com/golang/go/wiki/CodeReviewComments"}
```

Сервер должен ответить json'ом следующего вида:
```
{
  "url": "https://github.com/golang/go/wiki/CodeReviewComments",
  "key": "7758b4"
}
```

`7758b4` здесь - это сгенерированное сервисом число.

После такого `/shorten` можно делать `/go/7758b4`.
Ответ должен иметь HTTP код 302.
302 указывает на то, что запрошенный ресурс был временно перемещен на другой адрес (передаваемый в HTTP header'е `Location`).

Если открыть http://localhost:6029/go/7758b4 в браузере, тот перенаправит на https://github.com/golang/go/wiki/CodeReviewComments.

Сервер должен слушать порт, переданный через аргумент `-port`.

### Примеры

Запуск:
```
$ go run main.go -port 6029
```

Успешное добавление URL'а (200, Content-Type: application/json):
```
$ curl -i -X POST  "localhost:6029/shorten" -d '{"url":"https://github.com/golang/go/wiki/CodeReviewComments"}'
HTTP/1.1 200 OK
Content-Type: application/json
Date: Sat, 15 Feb 2020 23:35:26 GMT
Content-Length: 82

{"url":"https://github.com/golang/go/wiki/CodeReviewComments","key":"65ed150831"}
```

Невалидный json (400):
```
$ curl -i -X POST  "localhost:6029/shorten" -d '{"url":"https://github.com'                                   
HTTP/1.1 400 Bad Request
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Sat, 15 Feb 2020 23:30:27 GMT
Content-Length: 16

invalid request
```

Успешный запрос (302, Location header):
```
$ curl -i -X GET  "localhost:6029/go/c1464c853a"                                                               
HTTP/1.1 302 Found
Content-Type: text/html; charset=utf-8
Location: https://github.com/golang/go/wiki/CodeReviewComments
Date: Sat, 15 Feb 2020 23:25:26 GMT
Content-Length: 75

<a href="https://github.com/golang/go/wiki/CodeReviewComments">Found</a>.
```

Несуществующий key (404):
```
$ curl -i -X GET  "localhost:6029/go/uaaab"
HTTP/1.1 404 Not Found
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Sat, 15 Feb 2020 23:26:48 GMT
Content-Length: 14

key not found
```

### Состояние
Своё состояние сервис хранит в памяти.

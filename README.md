## Тех задача 

### Для начало теста нужно поменят значении в файле /env/.env

```sh
    DATABASE = "user=postgres password=123456 dbname=postgres port=5432 sslmode=disable"
    PORT = "8080"
  ```
### Чтобы начать программу 

  ```sh
    go run /cmd/server.go
  ```

### Недостатки 
  * Нужна добавлят тесты 
  * Необходима оптимизировать добавления и чтения записей бд
  * Нужен докер и докер компос для быстроты теста
  * На вебсокете лучще отправлять уведомление о новых данных чтобы клиенты сами сделали запрос на /flights

### Генерировать данные

  ```sh
  // https://json-generator.com/

  [
  '{{repeat(1000, 2000)}}',
  {
     Flight: '{{integer(1, 400)}}',
     From:'{{city()}}',
     Departure:'{{date(new Date(2014, 0, 1), new Date(), "YYYY-MM-ddThh:mm:ss")}}.511Z',
     To:'{{city()}}',
     Arrival:'{{date(new Date(2014, 0, 1), new Date(), "YYYY-MM-ddThh:mm:ss")}}.511Z'
  }
]
```
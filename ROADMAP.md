# План реализации

01-07 - MVP

## [01 - Simple storage](/docs/01-stage.md) ✅

Один узел, сохраняющий данные

### Архитектура

Client -> Gateway -> Node

### Требования

Поддерживаются операции:

- `PUT bucket/object`
- `GET bucket/object`
- `DELETE bucket/object`
- `GET bucket` (решено сделать этот эндпоинт на следующем этапе)

## [02 - Metadata Service](/docs/02-stage.md) ✅

Сервис, знающий о файлах. Промежуточный этап для хранения данных на нескольких
узлах. Освобождает Gateway от будущей логики распределения между узлами

### Архитектура

```
Client -> Gateway -> Node
              |
          Metadata Service
```

### Требования

Metadata Service хранит:

- bucket
- key
- size
- etag
- created at
- updated at
- storage node
- object path
- system metadata (map string -> string)
- user metadata (map string -> string)

## [03 - Несколько Storage Node](/docs/03-stage.md) ✅

Данные сохраняются на один из нескольких узлов. Metadata Service выбирает, на
какой из узлов попадет объект

### Архитектура

```
                  -> Node 1 
Client -> Gateway -> Node 2 
              |   -> Node 3 
          Metadata Service
```

### Требования

Round-Robin стратегия выбора узла

## [04 - Node discovery & Heartbeat](docs/04-stage.md) ✅

Каждый Storage Node раз в заданное время делает `heartbeat`, так что Metadata
Service всегда знает, кто жив

### Требования

Metadata Service дополнительно хранит для каждого живого узла:

- last seen
- node address

## 05 - Storage Node disk space

Вместе с `heartbeat` узел отдает доступное на его диске место.
Эта информация хранится в Metadata сервисе и будет использоваться для
выбора узлов для восстановления данных

### Требования

Metadata Service дополнительно хранит для каждого живого узла:

- free space bytes

## 06 - Replication

Данные сохраняются на несколько узлов (replication factor).
Metadata Service дополнительно хранит идентификатор узла с главной репликой.
Если узел с главной репликой умер, данные можно получить с других узлов.

### Требования

Выделяется узел с Primary Replica, на который будут приходить все запросы на
работу с данными.

## 07 - Recovery

При падении узла данные, хранящиеся на нем, копируются на другие живые узлы

### Архитектура

Metadata Service обнаруживает что Node умер и инициирует восстановление

## 08 - Streaming

Upload / download потоком, а не целым файлом

## 09 - Chunking

Данные хранятся не целиком, а по частям - чанкам

### Архитектура

Gateway делит полученный объект на части (чанки) и сохраняет каждый чанк как
отдельный объект

### Требования

Разные чанки лежат на разных узлах
Metadata Service хранит расположение всех чанков

## 10 - Consistent Hashing

При добавлении в кластер нового узла не нужно переносить все данные

## 11 - Background Rebalancing

Перенос данных делается в фоне по частям

## xx - Несколько Metadata Service

Система поддерживает несколько экземпляров Metadata Service, что позволяет
продолжать работу при падании одного из них  

## xx - Versioning

Система поддерживает версионирование объектов. Можно получать разные весии
одного объекта `GET bucket/object?version=3`

## xx - Compression

## xx - Monitoring

## xx - AAA

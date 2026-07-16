# Go-Storage
Распределенное объектное хранилище на Go, поддерживающее хранение файлов, репликацию между несколькими узлами и автоматическое восстановление после отказов

[Архитектурные ограничения](docs/architecture.md)

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

## 02 - Metadata Service
Сервис, знающий о файлах. Промежуточный этап для хранения данных на нескольких узлах. Освобождает Gateway от будущей логики распределения между узлами

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
- checksum 
- created at 
- updated at 
- storage node

## 03 - Несколько Storage Node 
Данные сохраняются на один из нескольких узлов. Metadata Service выбирает, на какой их узлов попадет объект

### Архитектура 
```
                  -> Node 1 
Client -> Gateway -> Node 2 
              |   -> Node 3 
          Metadata Service
```

### Требования 
Round-Robin стратегия выбора узла 

## 04 - Heartbeat
Каждый Storage Node раз в заданное время делает `POST /heartbeat`, так что Metadata Service всегда знает, кто жив
Metadata Service дополнительно хранит:
- last seen 
- free_space 

## 05 - Node Discovery
Storage Node может сделать `POST /register` на Metadata Service, чтобы начать получать данные

## 06 - Replication
Данные сохраняются на несколько узлов (replication factor). Промежуточный этап для логики восстановления
Metadata Service дополнительно хранит идентификатор узла с главной репликой

### Требования
Выделяется узел с Primary Replica, на который будут приходить все запросы на работу с данными 

## 07 - Recovery
При падении узла данные, хранящиеся на нем, копируются на другие живые узлы

### Архитектура
Metadata Service обнаруживает что Node умер и инициирует восстановление

## 08 - Streaming
Upload / download потоком, а не целым файлом 

## 09 - Chunking
Данные хранятся не целиком, а по частям - чанкам 

### Архитектура
Gateway делит полученный объект на части (чанки) и сохраняет каждый чанк как отдельный объект 

### Требования 
Разные чанки лежат на разных узлах 
Metadata Service хранит расположение всех чанков

## 10 - Consistent Hashing
При добавлении в кластер нового узла не нужно переносить все данные

## 11 - Background Rebalancing
Перенос данных делается в фоне по частям

## xx - Несколько Metadata Service
Система поддерживает несколько экземпляров Metadata Service, что позволяет продолжать работу при падании одного из них  

## xx - Versioning
Система поддерживает версионирование объектов. Можно получать разные весии одного объекта `GET bucket/object?version=3`

## xx - Compression

## xx - Monitoring


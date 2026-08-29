# Go Storage

*Go Storage* - это распределенное объектное хранилище, написанное на Go, с
поддержкой S3-подобного API.

> *Статус*: активная разработка. [Этапы](ROADMAP.md) 01-04 реализованы.

---

## Текущие возможности

На данный момент система поддерживает:

- Базовые операции работы с объектами: загрузка `PUT`, скачивание `GET`, удаление
`DELETE`, получение метаинфомрации объекта `HEAD`
- Базовые операции работы с бакетами: создание `PUT`, получение списка
доступных бакетов `GET`, удаление `DELETE`, получение метаинфомрации бакета `HEAD`
- Получение списка объектов в бакете (с поддержкой common prefixes, delimiter),
- Service Discovery & Health Check: автоматическое обнаружение новых Storage Node,
- Горизонтальное масштабирование: данные распределяются между разными Storage Node.

---

## Архитектура

*Gateway* - принимает HTTP запросы от клиентов, валидирует их и оркестрирует
работу.

*Metadata* - хранит информацию об объектах, ведет реестр живых узлов и выбирает,
куда загрузить новый объект.

*Storage Node* - непостредственно хранит данные объектов.

*Users* - сервис-заглушка для хранения списка зарегистрированных пользователей.

---

## Quick Start

1. Запуск кластера

```bash
git clone https://github.com/neelalala/go-storage.git
cd go-storage
docker compose up -d --build
```

Это поднимет 1 Gateway, 1 Metadata, 3 Storage Node, 1 PostgreSQL

1. Использование

a. Регистрация

> Зарегистрировать пользователя `new-user` в системе:

```bash
curl -X PUT localhost:8080/users/new-user
```

b. Работа с бакетами

> Создать бакет `new-bucket`:

```bash
curl -X PUT localhost:8080/storage/new-bucket \
-H 'Authorization: Username new-user'
```

> Получить список доступных пользователю бакетов:

```bash
curl localhost:8080/storage/ \
-H 'Authorization: Username new-user'
```

> Получить метаинфомрацию бакета:

```bash
curl -X HEAD localhost:8080/storage/new-bucket \
-H 'Authorization: Username new-user'
```

> Удалить пустой бакет:

```bash
curl -X DELETE localhost:8080/storage/new-bucket \
-H 'Authorization: Username new-user'
```

c. Работа с объектами

> Сохранть файл `my-photo.jpeg` с ключом `images/my-photo.jpeg` и доавить к метаинформации описание изображения:

```bash
curl -T ./my-photo.jpeg localhost:8080/storage/new-bucket/images/my-photo.jpeg \
-H 'Authorization: Username new-user' \
-H 'Content-Type: image/jpeg' \
-H 'X-Amz-Meta-Image-Description: My picture'
```

> Скачать объект `images/my-photo.jpeg` из хранилища:

```bash
curl -O localhost:8080/storage/new-bucket/images/my-photo.jpeg \
-H 'Authorization: Username new-user'
```

> Получить список объектов в бакете `new-bucket` с префиксом `images` и
разделителем `/` (имитация файловой системы: объекты в директории `images`):

```bash
curl 'localhost:8080/storage/new-bucket?prefix=images/&delimiter=/&limit=10&offset=0' \
-H 'Authorization: Username new-user'
```

> Получить метаинформацию объекта `images/my-photo.jpeg` (помимо системной,
также пользовательскую, заголовки начинающиеся с X-Amx-Meta):

```bash
curl -X HEAD localhost:8080/storage/new-bucket/images/my-photo.jpeg \
-H 'Authorization: Username new-user'
```

> Удалить объект `images/my-photo.jpeg`:

```bash
curl -X DELETE localhost:8080/storage/new-bucket/images/my-photo.jpeg \
-H 'Authorization: Username new-user'
```

---

## Технический стек

- *Язык*: Go
- *Внутренняя коммуникация*: gRPC и Protobuf
- *База данных для метаинфомрации*: PostgreSQL
- *Инфраструктура*: Docker, Docker compose

---

## [Roadmap](ROADMAP.md)

### MVP

- 01 - Simple Storage
- 02 - Metadata Service
- 03 - Несколько Storage Node
- 04 - Node Discovery & Heartbeat
- 05 - Storage Node disk space
- 06 - Replication
- 07 - Recovery

### Advanced

- 08 - Streaming
- 09 - Chunking
- 10 - Consistent Hashing
- 11 - Background Rebalancing

### Future

- Несколько Metadata Service
- Object Versioning
- Compression
- Monitoring & Metrics
- AAA

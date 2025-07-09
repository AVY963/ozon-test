# Ozon-test

Система для постов и комментариев с GraphQL API на Go.

## Описание

Классическая система постов с комментариями, как на Хабре или Reddit. Поддерживает создание постов, комментирование с неограниченной вложенностью, real-time уведомления.

### Основные возможности:
- **Пользователи**: создание, обновление, удаление, валидация email и username
- **Посты**: CRUD операции, отключение комментариев, пагинация, получение по автору
- **Комментарии**: иерархическая структура (materialized path), до 2000 символов, пагинация на всех уровнях
- **Real-time**: WebSocket подписки на новые комментарии к посту
- **Хранилище**: PostgreSQL и in-memory

## Запуск

### Быстрый старт с in-memory хранилищем
```bash
# Клонирование репозитория
git clone <repository-url>
cd ozon-posts

# Запуск с хранилищем в памяти
make run-memory
```

### Запуск с PostgreSQL
```bash
# Установка инструментов для миграций
make install

# Запуск PostgreSQL в Docker
make db-up

# Миграции базы данных
DB_URL="postgres://postgres:postgres@localhost:5432/ozon_posts?sslmode=disable" make migrate-up

# Запуск приложения
make run-postgres
```

### Запуск в Docker
```bash
# Сборка образа
make docker-build

# Запуск контейнера
make docker-run
```

### Остановка
```bash
# Остановка PostgreSQL
make db-down
```

## API

GraphQL API доступно на **http://localhost:8080/query**

Встроенный **GraphQL Playground** доступен по тому же адресу для интерактивного тестирования API.

## Что реализовано

### Queries
- `user(id: String!)` - получение пользователя по ID
- `userByUsername(username: String!)` - поиск по имени
- `posts(limit: Int, offset: Int)` - список постов с пагинацией  
- `postsByAuthor(authorId: String!)` - посты конкретного автора
- `post(id: String!)` - пост с комментариями
- `postComments(postId: String!)` - комментарии к посту
- `commentReplies(parentId: String!)` - ответы на комментарий
- `commentThread(commentId: String!, maxDepth: Int)` - цепочка комментариев

### Mutations  
- `createUser/updateUser/deleteUser` - управление пользователями
- `createPost/updatePost/deletePost` - управление постами
- `toggleComments` - включение/отключение комментариев к посту
- `createComment/updateComment/deleteComment` - управление комментариями

### Subscriptions
- `commentAdded(postId: String!)` - подписка на новые комментарии к посту

### Валидация
- **Username**: 3-50 символов, без пробелов
- **Email**: корректный формат
- **Пост**: заголовок до 200 символов, контент до 10000
- **Комментарий**: до 2000 символов

### Особенности
- **Materialized Path** для эффективной работы с иерархией комментариев
- **UUID** для всех сущностей
- **Graceful shutdown** с таймаутом 30 секунд
- **Логирование** через Logrus с JSON форматом
- **Проверка прав**: редактировать можно только свои посты/комментарии

## Тестирование

```bash
# Все тесты
make test

# Покрытие кода
./scripts/run_tests.sh
```

Тесты включают:
- Юнит-тесты сущностей и сервисов
- Интеграционные тесты полного workflow
- Тесты системы ошибок

## Конфигурация

Настройка через переменные окружения:

```bash
# Тип базы данных
DB_TYPE=postgres # или memory

# PostgreSQL (если DB_TYPE=postgres)
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=ozon_posts

# Сервер
PORT=8080
HOST=0.0.0.0

# Логирование
LOG_LEVEL=info
LOG_FORMAT=json
```

## Архитектура

```
internal/
├── config/          # Конфигурация приложения
├── entities/        # Доменные сущности
├── services/        # Бизнес-логика  
├── repositories/    # Слой доступа к данным
│   ├── inmemory/   # In-memory реализации
│   └── postgres/   # PostgreSQL реализации
└── handlers/        # HTTP handlers
    └── graphql/    # GraphQL resolvers

pkg/
├── errors/         # Система ошибок
├── logger/         # Настройка логирования  
└── testutils/      # Утилиты для тестов
```
 
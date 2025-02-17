# Merch Store Service

## Обзор
В Авито существует внутренний магазин мерча, где сотрудники могут приобретать товары за монеты (coin). Каждому новому сотруднику выделяется 1000 монет, которые можно использовать для покупки товаров. Кроме того, монеты можно передавать другим сотрудникам в знак благодарности или как подарок.

## Основные Функции

1. Обмен монетами между сотрудниками.
2. Покупка мерча за монеты.
3. Управление балансом.
4. Отслеживание транзакций.

## Установка и настройка

### Требования
- Go 1.22+
- Docker и Docker Compose
- PostgreSQL

### Запуск приложения
```sh
make run
```
Или с использованием Docker:
```sh
docker compose up --build
```

### Запуск миграций
```sh
make migrate-up
make migrate-down
make migrate-force
make migrate-version
```

### Запуск тестов
```sh
make runtest
```

## API эндпоинты

### Аутентификация
- `POST /api/auth` - Авторизация пользователя и выдача JWT-токена.

### Получение информации о пользователе
- `GET /api/info` - Позволяет получить сведения о доступных монетах, инвентаре и истории транзакций пользователя.
- `POST /api/sendCoin` - Позволяет отправить определенное количество монет другому пользователю.

### Операции магазина
- `GET /api/buy/{item}` - Позволяет купить предмет, используя монеты пользователя.

## Конфигурация
Конфигурация управляется через YAML-файлы в каталоге `configs/`.

| Переменная  | Описание           |
|-------------|--------------------|
| `dbhost`    | Хост базы данных   |
| `dbport`    | Порт базы данных   |
| `dbname` | Имя базы данных    |
| `jwtsecret` | Секретный ключ JWT |



# Княжества Древней Руси: бэкенд

Курс «Разработка интернет-приложений», МГТУ им. Баумана, ИУ5-52Б, осень 2026. Тема «История и культура, вариант 2»: расчёты исторических данных по населению на основе археологии.

Система хранит удельные княжества Древней Руси с фотографией и коротким видео их укреплений. Археолог заводит черновик княжества, дописывает описание раскопок, дату основания и коэффициент земли, публикует его, ставит лайки чужим княжествам и убирает своё, если ошибся. Лента показывает опубликованные княжества по одному, плитка даёт список с фильтром по дате основания.

Репозитории проекта:

- бэкенд: этот репозиторий;
- фронтенд: [principality-frontend](https://github.com/dehwyy/principality-frontend) (React, лабораторная 5);
- мобильное приложение и асинхронный сервис: появятся с домашним заданием.

## Стек

- Go 1.26, gin, GORM;
- PostgreSQL 16, Adminer;
- Minio для изображений и видео;
- html/template для страниц лабораторных 1 и 2;
- JWT (`golang-jwt/jwt`, HS256) для аутентификации археологов, Redis 6.2 для blacklist токенов после выхода;
- Swagger UI (`swaggo/swag`, `gin-swagger`) для описания методов.

## Запуск

```bash
docker compose up -d

docker exec -it minio_storage mc alias set myminio http://localhost:9000 root rootpassword
docker exec -it minio_storage mc mb myminio/principality-media
docker exec -it minio_storage mc anonymous set public myminio/principality-media

cp .env.example .env
go run ./cmd/migrate
docker exec -i principality_db psql -U principality -d principality < migrations/principalities_seed.sql
go run ./cmd/principality-web
```

Медиа десяти княжеств из сида кладутся в бакет `principality-media` под именами из столбцов `image_key` и `video_key` (`kiev.jpg`, `kiev.mp4` и так далее), например через `mc cp`. Сервер слушает `http://localhost:8080`, Adminer доступен на `http://localhost:8081`, консоль Minio на `http://localhost:9001`.

В `.env` задаются подключение к PostgreSQL (`DB_*`), Minio (`MINIO_ENDPOINT`, `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`, `MINIO_BUCKET_NAME`), Redis (`REDIS_HOST`, `REDIS_PORT`, `REDIS_USER`, `REDIS_PASSWORD`) и секрет подписи JWT (`JWT_TOKEN`). Адрес, по которому клиент получает файлы, срок жизни токена `[JWT] ExpiresIn` и таймауты `[Redis]` лежат в `config/config.toml`.

Swagger UI доступен на `http://localhost:8080/swagger/index.html`. Описание генерируется командой `swag init -g cmd/principality-web/main.go -o docs`.

## HTTP-методы

Все методы веб-сервиса начинаются с `/api`. Гостю доступны список, лента, регистрация и аутентификация. Остальные методы требуют заголовок `Authorization: Bearer <access_token>`; без него, с токеном из blacklist или с неверной подписью методы создателя отвечают 403 с пустым телом, выход без токена отвечает 400. Создатель княжества берётся из токена. Остальные ошибки имеют вид `{"status": "error", "description": "..."}`. Даты в ответах в формате RFC 3339. Княжества в статусе `removed` наружу не отдаются ни одним методом.

### Домен княжества

| Метод | URL | Назначение | Вход | Ответ | Коды |
|---|---|---|---|---|---|
| GET | `/api/principalities?foundedBefore=ГГГГ-ММ-ДД` | Список опубликованных княжеств с фильтром по дате основания | query `foundedBefore`, необязательный | Массив княжеств с URL медиа, числом лайков и признаком `created_by_current_archaeologist` (0/1) | 200, 400, 500 |
| GET | `/api/principalities/feed`, `/api/principalities/feed/{principalityId}?next=true` | Лента: первое опубликованное княжество, княжество по ид или следующее после него; после последнего лента начинается с первого | path `principalityId`, query `next` | Княжество с URL медиа, числом лайков и признаком `principality_liked` (0/1) | 200, 400, 404, 500 |
| GET | `/api/principalities/draft` | Черновик текущего археолога | нет | Княжество-черновик с URL медиа | 200, 403, 404, 500 |
| POST | `/api/principalities` | Создание черновика с изображением и видео | multipart: `principality_name`, файлы `principality_image`, `principality_video` | Созданный черновик; ключи файлов в `image_key` и `video_key` сгенерированы на латинице | 201, 400, 403, 409, 500 |
| PUT | `/api/principalities/draft/publish` | Публикация своего черновика, ид не указывается | JSON: `principality_summary`, `founding_date` (`ГГГГ-ММ-ДД`), `land_coefficient` (от 0 до 1) | Опубликованное княжество | 200, 400, 403, 404, 409, 500 |
| DELETE | `/api/principalities/{principalityId}` | Логическое удаление своего княжества | нет | `{"status": "success", "message": "княжество удалено"}` | 200, 400, 403, 404, 500 |
| POST | `/api/principalities/{principalityId}/like` | Лайк от текущего археолога | JSON: `principality_liked` (1 ставит, 0 снимает) | `principality_id`, `principality_liked`, `principality_like_count` | 200, 400, 403, 404, 500 |

### Домен археолога

| Метод | URL | Назначение | Вход | Ответ | Коды |
|---|---|---|---|---|---|
| POST | `/api/archaeologists` | Регистрация археолога | JSON: `archaeologist_login`, `archaeologist_password` | `archaeologist_id`, `archaeologist_login`, `is_chronicle_keeper`; пароль не возвращается | 201, 400, 409, 500 |
| POST | `/api/archaeologists/login` | Аутентификация, выдача JWT (HS256, срок из конфига) | JSON: `archaeologist_login`, `archaeologist_password` | `expires_in`, `access_token`, `token_type` (`Bearer`) | 200, 400, 403, 500 |
| POST | `/api/archaeologists/logout` | Деавторизация: токен кладётся в blacklist Redis под ключом `principality.jwt.<токен>` на срок жизни токена | заголовок `Authorization: Bearer <access_token>` | пустой ответ | 200, 400, 500 |

### Объект княжества в ответах

Черновик, создание и публикация возвращают базовый объект. Список добавляет `created_by_current_archaeologist` и `principality_like_count`, лента добавляет `principality_like_count` и `principality_liked`.

| Поле | Тип | Смысл |
|---|---|---|
| `principality_id` | число | Идентификатор |
| `principality_name` | строка | Название |
| `principality_summary` | строка | Описание раскопок, у черновика пустое |
| `principality_status` | строка | `draft` или `published` |
| `image_key`, `video_key` | строка | Имена файлов в Minio |
| `principality_image_url`, `principality_video_url` | строка | Полные URL медиа для клиента |
| `founding_date` | строка RFC 3339 | Дата основания, у черновика `null` |
| `land_coefficient` | число | Коэффициент земли, у черновика `null` |
| `created_at`, `published_at` | строка RFC 3339 | Создание и публикация, `published_at` у черновика `null` |
| `created_by` | число | Идентификатор археолога-создателя |
| `created_by_current_archaeologist` | 0 или 1 | Создал ли княжество текущий археолог, только в списке |
| `principality_like_count` | число | Сколько лайков, в списке и ленте |
| `principality_liked` | 0 или 1 | Лайкнул ли княжество текущий археолог, только в ленте |

### Страницы на шаблонах (лабораторные 1 и 2)

| Метод | URL | Страница |
|---|---|---|
| GET | `/principalities/feed/{principalityId}?next=true` | Лента |
| GET | `/principalities/draft` | Добавление княжества |
| GET | `/principalities?foundedBefore=ГГГГ-ММ-ДД` | Плитка с фильтром |
| POST | `/principalities/draft` | Создание черновика из формы |
| POST | `/principalities/{principalityId}/publish` | Публикация из формы |
| POST | `/principalities/{principalityId}/remove` | Удаление кнопкой на плитке сырым `UPDATE` |

## Таблицы базы данных

### `principality`, княжества

| Столбец | Тип | Ограничения | Смысл |
|---|---|---|---|
| `principality_id` | `BIGINT` | PK | Идентификатор княжества |
| `principality_name` | `VARCHAR(120)` | NOT NULL | Название княжества |
| `principality_summary` | `VARCHAR(600)` | | Что известно об укреплениях и застройке по раскопкам |
| `principality_status` | `VARCHAR(16)` | NOT NULL, по умолчанию `draft` | `draft`, `published` или `removed` |
| `image_key` | `VARCHAR(80)` | NOT NULL | Имя файла изображения в Minio |
| `video_key` | `VARCHAR(80)` | NOT NULL | Имя файла видео в Minio |
| `founding_date` | `DATE` | | Дата основания княжества, по ней фильтруется список |
| `land_coefficient` | `NUMERIC(4,2)` | | Доля укреплённой площади под усадебной застройкой |
| `created_at` | `TIMESTAMP WITH TIME ZONE` | NOT NULL | Когда заведён черновик |
| `published_at` | `TIMESTAMP WITH TIME ZONE` | | Когда княжество опубликовано |
| `created_by` | `BIGINT` | NOT NULL, FK на `archaeologist` без каскада | Археолог, заведший княжество |

Частичный уникальный индекс `ux_principality_one_draft_per_archaeologist` по `created_by` при `principality_status = 'draft'` не даёт археологу завести второй черновик.

Переходы статусов: `draft` переходит в `published` при публикации, `draft` и `published` переходят в `removed` при удалении. Обратно в черновик вернуться нельзя, удалённое княжество не восстанавливается.

### `archaeologist`, археологи

| Столбец | Тип | Ограничения | Смысл |
|---|---|---|---|
| `archaeologist_id` | `BIGINT` | PK | Идентификатор археолога |
| `archaeologist_login` | `VARCHAR(25)` | NOT NULL, UNIQUE | Логин |
| `archaeologist_password` | `VARCHAR(100)` | NOT NULL | SHA-1 пароля в шестнадцатеричном виде |
| `is_chronicle_keeper` | `BOOLEAN` | по умолчанию `false` | Признак хранителя летописи, попадает в роль внутри JWT |

### `principality_like`, лайки

| Столбец | Тип | Ограничения | Смысл |
|---|---|---|---|
| `principality_like_id` | `BIGINT` | PK | Идентификатор лайка |
| `archaeologist_id` | `BIGINT` | NOT NULL, FK на `archaeologist` без каскада | Кто поставил лайк |
| `principality_id` | `BIGINT` | NOT NULL, FK на `principality` без каскада | Какому княжеству |

Пара `archaeologist_id`, `principality_id` уникальна (`idx_principality_like`): один археолог ставит одному княжеству не больше одного лайка. Снятие лайка удаляет строку.

## Правила бизнес-логики

- Текущий археолог в методах API берётся из JWT: `POST /api/archaeologists/login` выдаёт токен, middleware `WithAuthCheck` проверяет заголовок `Authorization: Bearer`, blacklist в Redis, подпись и срок токена и кладёт идентификатор археолога в контекст запроса. Без токена доступны только список и лента, признаки `created_by_current_archaeologist` и `principality_liked` в них равны 0; методы создателя отвечают 403. Страницы лабораторных 1 и 2 по-прежнему работают от константного археолога `a.kuza` из функции-singleton `auth.CurrentArchaeologist()`.
- Системные поля с клиента не принимаются: идентификаторы, статусы, создатель, даты создания и публикации вычисляются на сервере. Во входных структурах их нет.
- Удалять можно только своё княжество, иначе 403. Публикуется только свой черновик: ид в запросе нет, сервер берёт черновик текущего археолога, а если черновика нет, отвечает 404.
- Лайк ставится только опубликованному княжеству. Повторный лайк не дублирует строку.
- Файлы уходят в Minio до записи в базу; имена генерируются как `principality_` плюс 16 шестнадцатеричных символов и расширение по реальному типу содержимого. Имя файла клиента не используется.
- Список, лента и черновик показывают только то, что положено: опубликованные княжества и черновик текущего археолога.

## Документация

- `docs/principalities.mdj`: ER-диаграмма, диаграмма классов и диаграмма последовательности в StarUML.
- `docs/principalities.postman_collection.json`: коллекция из десяти запросов к веб-сервису.
- `docs/swagger.json`, `docs/swagger.yaml`, `docs/docs.go`: описание API для Swagger, сгенерировано `swag`.
- `migrations/principalities_seed.sql`: десять княжеств, пять археологов и лайки для демонстрации.

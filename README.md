# Княжества Древней Руси: бэкенд

Курс «Разработка интернет-приложений», МГТУ им. Баумана, ИУ5-52, осень 2026. Тема «История и культура, вариант 2»: расчёты исторических данных по населению на основе археологии.

Система хранит удельные княжества Древней Руси с фотографией и коротким видео их укреплений. Археолог заводит черновик княжества, дописывает описание раскопок, дату основания и коэффициент застроенной земли, публикует его, ставит лайки чужим княжествам и убирает своё, если ошибся. Лента показывает опубликованные княжества по одному, плитка даёт список с фильтром по дате основания.

Репозитории проекта:

- бэкенд: этот репозиторий;
- фронтенд: появится с лабораторной 5;
- мобильное приложение и асинхронный сервис: появятся с домашним заданием.

## Стек

- Go 1.26, gin, GORM;
- PostgreSQL 16, Adminer;
- Minio для изображений и видео;
- html/template для страниц лабораторных 1 и 2;
- Redis подключится с лабораторной 4 вместе с JWT.

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

В `.env` задаются подключение к PostgreSQL (`DB_*`) и Minio (`MINIO_ENDPOINT`, `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`, `MINIO_BUCKET_NAME`). Адрес, по которому клиент получает файлы, лежит в `config/config.toml`.

## HTTP-методы

Все методы веб-сервиса начинаются с `/api`. Ответ с ошибкой всегда имеет вид `{"status": "error", "description": "..."}`. Даты в ответах в формате RFC 3339. Княжества в статусе `removed` наружу не отдаются ни одним методом.

### Домен княжества

| Метод | URL | Назначение | Вход | Ответ | Коды |
|---|---|---|---|---|---|
| GET | `/api/principalities?foundedBefore=ГГГГ-ММ-ДД` | Список опубликованных княжеств с фильтром по дате основания | query `foundedBefore`, необязательный | Массив княжеств с URL медиа, числом лайков и признаком `created_by_current_archaeologist` (0/1) | 200, 400, 500 |
| GET | `/api/principalities/feed`, `/api/principalities/feed/{principalityId}?next=true` | Лента: первое опубликованное княжество, княжество по ид или следующее после него | path `principalityId`, query `next` | Княжество с URL медиа, числом лайков и `next_principality_id` | 200, 400, 404, 500 |
| GET | `/api/principalities/draft` | Черновик текущего археолога | нет | Княжество-черновик с URL медиа | 200, 404, 500 |
| POST | `/api/principalities` | Создание черновика с изображением и видео | multipart: `principality_name`, файлы `principality_image`, `principality_video` | Созданный черновик; ключи файлов в `image_key` и `video_key` сгенерированы на латинице | 201, 400, 409, 500 |
| PUT | `/api/principalities/{principalityId}/publish` | Публикация своего черновика | JSON: `principality_summary`, `founding_date` (`ГГГГ-ММ-ДД`), `land_coefficient` (от 0 до 1) | Опубликованное княжество | 200, 400, 403, 404, 409, 500 |
| DELETE | `/api/principalities/{principalityId}` | Логическое удаление своего княжества | нет | `{"status": "success", "message": "княжество удалено"}` | 200, 400, 403, 404, 500 |
| POST | `/api/principalities/{principalityId}/like` | Лайк от текущего археолога | JSON: `principality_liked` (1 ставит, 0 снимает) | `principality_id`, `principality_liked`, `principality_like_count` | 200, 400, 404, 500 |

### Домен археолога

| Метод | URL | Назначение | Вход | Ответ | Коды |
|---|---|---|---|---|---|
| POST | `/api/archaeologists` | Регистрация археолога | JSON: `archaeologist_login`, `archaeologist_password` | `archaeologist_id`, `archaeologist_login`, `is_chronicle_keeper`; пароль не возвращается | 201, 400, 409, 500 |
| POST | `/api/archaeologists/login` | Аутентификация, заглушка до лабораторной 4 | JSON: `archaeologist_login`, `archaeologist_password` | `{"status": "success", "message": "..."}` | 200, 400 |
| POST | `/api/archaeologists/logout` | Деавторизация, заглушка до лабораторной 4 | нет | `{"status": "success", "message": "..."}` | 200 |

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
| `archaeologist_password` | `VARCHAR(100)` | NOT NULL | Пароль, до лабораторной 4 хранится как прислан |
| `is_chronicle_keeper` | `BOOLEAN` | по умолчанию `false` | Хранитель летописи, роль модератора |

### `principality_like`, лайки

| Столбец | Тип | Ограничения | Смысл |
|---|---|---|---|
| `principality_like_id` | `BIGINT` | PK | Идентификатор лайка |
| `archaeologist_id` | `BIGINT` | NOT NULL, FK на `archaeologist` без каскада | Кто поставил лайк |
| `principality_id` | `BIGINT` | NOT NULL, FK на `principality` без каскада | Какому княжеству |

Пара `archaeologist_id`, `principality_id` уникальна (`idx_principality_like`): один археолог ставит одному княжеству не больше одного лайка. Снятие лайка удаляет строку.

## Правила бизнес-логики

- Текущий археолог во всех методах один и тот же: `a.kuza` с идентификатором 1. Он задан константой в функции-singleton `auth.CurrentArchaeologist()`, которая в лабораторной 4 начнёт читать его из JWT.
- Системные поля с клиента не принимаются: идентификаторы, статусы, создатель, даты создания и публикации вычисляются на сервере. Во входных структурах их нет.
- Публиковать и удалять можно только своё княжество, иначе 403. Публикуется только черновик, иначе 409.
- Лайк ставится только опубликованному княжеству. Повторный лайк не дублирует строку.
- Файлы уходят в Minio до записи в базу; имена генерируются как `principality_` плюс 16 шестнадцатеричных символов и расширение по реальному типу содержимого. Имя файла клиента не используется.
- Список, лента и черновик показывают только то, что положено: опубликованные княжества и черновик текущего археолога.

## Документация

- `docs/principalities.mdj`: ER-диаграмма и диаграмма классов в StarUML.
- `docs/principalities.postman_collection.json`: коллекция из десяти запросов к веб-сервису.
- `migrations/principalities_seed.sql`: десять княжеств, пять археологов и лайки для демонстрации.

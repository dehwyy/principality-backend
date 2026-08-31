# Развёртывание `principality-backend`

Приложение живёт на `https://api.lab.14dev.ru`, в кластере `twc-test-cluster`,
в неймспейсе `availability`. Медиа отдаются тем же доменом по пути `/principality-media/`.

## Предусловия

```bash
export KUBECONFIG=~/Downloads/twc-test-cluster-config.yaml
kubectl get principalities                       # одна нода worker-192.168.0.5, amd64
gh auth status | grep scopes            # среди скоупов обязан быть write:packages
```

Если `write:packages` нет:

```bash
gh auth refresh -s write:packages -h github.com
docker logout ghcr.io
gh auth token | docker login ghcr.io -u dehwyy --password-stdin
```

`docker logout` здесь обязателен. Docker Desktop держит токен в системном
credsStore и после обновления скоупов продолжает отдавать старый — пуш падает
с `403 Forbidden` на `POST /v2/.../blobs/uploads/`, хотя `docker login` пишет
`Login Succeeded`.

## Обычный передеплой

Три команды. Тег `latest` мутабельный, у пода `imagePullPolicy: Always`,
поэтому `rollout restart` подтягивает свежий образ.

```bash
cd principality-backend
docker buildx build --platform linux/amd64 -t ghcr.io/dehwyy/principality-backend:latest --push .
kubectl -n availability rollout restart deploy/principality-web
kubectl -n availability rollout status deploy/principality-web --timeout=180s
```

`--platform linux/amd64` не опция: Mac собирает arm64, нода — amd64.
Без флага под встанет в `CrashLoopBackOff` с `exec format error`.

Если менялись манифесты — перед рестартом:

```bash
kubectl apply -f deploy/
```

`kubectl` читает из каталога только `*.yaml`/`*.yml`/`*.json`, так что этот
файл и `upload-principality-media.sh` он игнорирует.

## Проверка после выката

```bash
for u in \
  https://api.lab.14dev.ru/principalities/feed/1 \
  https://api.lab.14dev.ru/principalities/draft \
  'https://api.lab.14dev.ru/principalities?minArea=99.9' \
  https://api.lab.14dev.ru/principalities/feed/999 \
  https://api.lab.14dev.ru/static/css/principalities.css \
  https://api.lab.14dev.ru/principality-media/ec2_single.jpg \
  https://api.lab.14dev.ru/principality-media/ec2_single.mp4 ; do
  curl -s -o /dev/null -w '%{http_code}  %{content_type}  %{url_effective}\n' "$u"
done
```

Ожидается: три `200 text/html`, `404 text/html` на несуществующий узел,
`200 text/css`, `200 image/jpeg`, `200 video/mp4`.

Отдельно — что в HTML стоят абсолютные URL Minio, иначе скриншоты 13–17 не снять:

```bash
curl -s https://api.lab.14dev.ru/principalities | grep -o 'src="[^"]*principality-media[^"]*"' | head -3
```

Фильтр по доступности (ожидается 8 / 5 / 3 / 0 карточек):

```bash
for v in 0 99.9 99.95 99.99; do
  echo -n "minArea=$v -> "
  curl -s "https://api.lab.14dev.ru/principalities?minArea=$v" | grep -c 'class="principality-card"'
done
```

## Медиа

Бакет `principality-media` лежит в **чужом** Minio — `aioffice/minio`. Скрипт создаёт
бакет, открывает анонимное чтение и заливает файлы из `../media/principalities`
с правильными `Content-Type`. Повторный запуск безопасен.

```bash
./deploy/upload-principality-media.sh
```

Скрипт поднимает временный под с `minio/mc` и льёт файлы через `mc pipe`.
Через `kubectl cp` не выйдет: в образе `minio/mc` нет `tar`, и копирование
падает с `exec: "tar": executable file not found in $PATH`.

**Проверять накануне каждой защиты.** Бакет живёт в неймспейсе постороннего
проекта; пересоздадут тот StatefulSet — медиа исчезнут вместе с ним.
Восстановление — тот же скрипт, исходники в `../media/principalities`.

## Карта ресурсов

| Что | Где |
|---|---|
| Deployment + Service `principality-web` | ns `availability`, порт 8080 |
| Service `principality-media` | ns `availability`, `ExternalName` → `minio.aioffice.svc.cluster.local:9000` |
| Ingress `principality-web` | `api.lab.14dev.ru`, класс `nginx`; `/principality-media` → Minio, `/` → приложение |
| Сертификат | secret `principality-web-tls`, ClusterIssuer `letsencrypt-prod` |
| Pull-secret | `ghcr-pull` в ns `availability` |
| Бакет медиа | `principality-media` в `aioffice/minio`, анонимное чтение |
| Образ | `ghcr.io/dehwyy/principality-backend`, теги `latest` и датированный |

Почему один домен на приложение и медиа: `*.lab.14dev.ru` не резолвится,
A-запись есть только у `api.lab.14dev.ru`. Отдельный поддомен под Minio
потребовал бы новой записи в DNS. Заодно нет mixed content: страница и медиа
идут по одному HTTPS.

Переменные окружения приложения задаются в `principality-web.yaml`:
`MINIO_BASE_URL` (схема и хост, без бакета — бакет подставляет шаблон),
`SERVER_ADDR`, `GIN_MODE`.

## Диагностика

| Симптом | Причина | Что делать |
|---|---|---|
| `403 Forbidden` при пуше | credsStore отдаёт старый токен | `docker logout ghcr.io` и залогиниться заново |
| `ImagePullBackOff`, `401 Unauthorized` | протух токен в `ghcr-pull` | пересоздать секрет (команда ниже) |
| `CrashLoopBackOff`, `exec format error` | образ собран под arm64 | пересобрать с `--platform linux/amd64` |
| Страницы открываются, медиа `404` | пропал бакет или объекты | `./deploy/upload-principality-media.sh` |
| Медиа `503` | Minio в `aioffice` лежит | `kubectl -n aioffice get pods -l app=minio` |
| Сертификат `Ready=False` дольше пары минут | ACME не проходит | `kubectl -n availability describe order` |

Пересоздать pull-secret:

```bash
kubectl -n availability create secret docker-registry ghcr-pull \
  --docker-server=ghcr.io --docker-username=dehwyy --docker-password="$(gh auth token)" \
  --dry-run=client -o yaml | kubectl apply -f -
kubectl -n availability rollout restart deploy/principality-web
```

Логи и события:

```bash
kubectl -n availability logs deploy/principality-web --tail=50
kubectl -n availability describe pod -l app=principality-web | tail -30
kubectl -n availability get events --sort-by=.lastTimestamp | tail -20
```

Откат на предыдущую ревизию:

```bash
kubectl -n availability rollout undo deploy/principality-web
```

Про сертификаты: у Let's Encrypt лимит 5 неудачных попыток на домен в час.
Не удалять и не пересоздавать Ingress «на пробу» — при неудаче разбираться
через `describe order`, а не повторными применениями.

## Развёртывание с нуля

```bash
export KUBECONFIG=~/Downloads/twc-test-cluster-config.yaml
kubectl apply -f deploy/namespace.yaml
kubectl -n availability create secret docker-registry ghcr-pull \
  --docker-server=ghcr.io --docker-username=dehwyy --docker-password="$(gh auth token)" \
  --dry-run=client -o yaml | kubectl apply -f -
docker buildx build --platform linux/amd64 -t ghcr.io/dehwyy/principality-backend:latest --push .
kubectl apply -f deploy/
./deploy/upload-principality-media.sh
kubectl -n availability rollout status deploy/principality-web --timeout=180s
```

Сертификат выпускается сам, обычно за 20–40 секунд:

```bash
kubectl -n availability get certificate -w
```

## Что добавится дальше

- **Лаба 2** — PostgreSQL. В кластере есть чужие экземпляры (`acchi`, `lootbox`),
  но под лабу нужен свой: Deployment + PVC на `local-path` в ns `availability`,
  DSN через Secret. Плюс миграции при старте пода.
- **Лаба 4** — Redis под чёрный список JWT. Deployment без тома, ClusterIP.
- **Лаба 5** — фронтенд отдельным репозиторием и отдельным доменом; понадобится
  новая A-запись, потому что wildcard на `lab.14dev.ru` нет.

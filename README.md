# Programowanie aplikacji mobilnych i webowych - projekt

Link do aplikacji uruchomionej na Heroku: <https://boiling-refuge-64939.herokuapp.com/>

## Zmienne środowiskowe

Aplikacja przy uruchomieniu wczytuje zmienne środowiskowe z pliku `.env`. Można w nim ustawić następujące parametry:
- `REDIS_HOST` - adres bazy danych redis, domyślna wartość `localhost`
- `REDIS_PORT` - port bazy danych redis, domyślna wartość `6379`
- `REDIS_PASS` - hasło do bazy danych redis
- `REDIS_DB` - numer bazy danych redis, domyślna wartość `0`
- `PORT` - port na którym ma zostać uruchomiona aplikacja, domyślna wartość `5000`
- `SECURE_COOKIE` - decyduje jaką wartość powinna mieć flaga `Secure` ciasteczek sesji, domyślna wartość `FALSE`
- `SESSION_NAME` - decyduje jaką nazwę powinny mieć ciasteczka sesji, domyślna wartość `session`
